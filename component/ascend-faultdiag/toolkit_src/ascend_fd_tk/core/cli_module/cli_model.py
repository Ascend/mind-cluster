#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import asyncio
import inspect
import os.path
import shutil

from ascend_fd_tk.core.cli_module.base import CliModel, DetailedCliModel, CliCtx
from ascend_fd_tk.core.common import diag_enum
from ascend_fd_tk.core.common.diag_enum import Customer
from ascend_fd_tk.core.common.errors import GenerateCsvPermissionErr
from ascend_fd_tk.core.common.path import CommonPath
from ascend_fd_tk.core.context.diag_ctx import DiagCtx
from ascend_fd_tk.examples.auto_diag.auto_collect import AutoCollect
from ascend_fd_tk.examples.auto_diag.auto_diag import AutoDiagCluster
from ascend_fd_tk.examples.auto_diag.auto_single_diag import AutoSingleDiag
from ascend_fd_tk.examples.auto_diag.collect_bmc_log import CollectBmcLog
from ascend_fd_tk.examples.inspection.inspection import Inspection
from ascend_fd_tk.utils import logger
from ascend_fd_tk.utils.file_tool import convert_log_path

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER


class HelpCliModel(CliModel):
    _SPACE_SIZE = 6

    @classmethod
    def get_key(cls) -> str:
        return "help"

    def get_help(self) -> str:
        return "显示帮助信息"

    def run_task(self, *args) -> str:
        results = []
        max_key_len = len(max(self.cli_ctx.cli_model_map.keys(), key=len))
        left_len = max_key_len + self._SPACE_SIZE
        for key, cli_model in self.cli_ctx.cli_model_map.items():
            results.append(f"{key:<{left_len}}- {cli_model.get_help()}")
        return "\n".join(results)


class ExitCliModel(CliModel):
    @classmethod
    def get_key(cls) -> str:
        return "exit"

    def get_help(self) -> str:
        return "退出程序"

    def run_task(self, *args) -> str:
        self.cli_ctx.is_running = False
        return "再见!"


class ClearCliModel(CliModel):
    @classmethod
    def get_key(cls) -> str:
        return "clear"

    def get_help(self) -> str:
        return "清屏"

    def run_task(self, *args) -> str:
        os.system('cls' if os.name == 'nt' else 'clear')  # nosec B605  硬编码命令，无注入风险
        return ""


class AboutCliModel(CliModel):
    @classmethod
    def get_key(cls) -> str:
        return "about"

    def get_help(self) -> str:
        return "查看关于诊断工具"

    def run_task(self, *args) -> str:
        return f"""
        MindCluster ascend-faultdiag-toolkit诊断工具版本：{self.diag_ctx.tool_config.version}
        """


class GuideCliModel(CliModel):
    @classmethod
    def get_key(cls) -> str:
        return "guide"

    def get_help(self) -> str:
        return "获取向导信息"

    def run_task(self, *args) -> str:
        return f"""
        一、采集内容准备
        请根据故障设备自行按需选择要采集的设备信息或需要导入的日志，可以不导入全量设备信息或日志。按需设置以下在线或离线采集分析的任意地址。

        1. 在线采集准备
        若需要在线采集设备信息，请使用 " {SetConnConfigCliModel.get_key()} " 命令设置设备信息，具体配置可使用" {SetConnConfigCliModel.get_key()} ? "查看详情

        2. 离线日志解析准备
        2.1 设置服务器日志目录地址
        请使用 " {SetHostDumpLogDirCliModel.get_key()} " 命令设置离线日志目录，具体配置可使用" {SetHostDumpLogDirCliModel.get_key()} ? "查看详情

        2.2 设置BMC日志目录地址
        请使用 " {SetBmcDumpLogDirCliModel.get_key()} " 命令设置离线日志目录，具体配置可使用" {SetBmcDumpLogDirCliModel.get_key()} ? "查看详情

        2.3 设置交换机回显文本目录地址
        请使用 " {SetSwiDumpLogDirCliModel.get_key()} " 命令设置离线日志目录，具体配置可使用"{SetSwiDumpLogDirCliModel.get_key()} ? "查看详情

        3. 默认读取路径
        当未手动设置以上文件或目录时，工具会自动读取执行路径下的以下默认文件或目录，相关文件或目录需用户提前手动创建：
        连接配置: conn.ini
        BMC日志目录: bmc_dump_log
        Host日志目录: host_dump_log
        交换机日志目录: switch_dump_log

        二、启动采集/分析 & 诊断
        执行 " {AutoCollectDiagCliModel.get_key()} " 启动在线采集/离线分析并诊断

        三、清理缓存
        本工具支持分批采集统一诊断，所以单次诊断完后会留有缓存，若已完成诊断任务，请使用 " {ClearCacheCliModel.get_key()} " 清理缓存(若无法有效清理，请使用管理员模式打开工具)，避免影响下次诊断结果

        总结:
        1. 先用 " {SetConnConfigCliModel.get_key()} " 设置要访问的设备ip配置文件或用 " {SetBmcDumpLogDirCliModel.get_key()} "，" {SetHostDumpLogDirCliModel.get_key()} "，" {SetSwiDumpLogDirCliModel.get_key()} "设置离线日志目录，或直接将日志放到默认目录下
        2. 以上至少有一项设置存在即可使用 " {AutoCollectDiagCliModel.get_key()} " 采集/分析并诊断输出报告
        """


class SetConfigDirCliModel(DetailedCliModel):
    @staticmethod
    def is_support_param():
        return True

    @staticmethod
    def check_input_path(*args) -> str:
        if not args:
            return "目录路径为空，请重新设置"
        dir_path = convert_log_path(args[0])
        if not dir_path:
            return "目录%s不存在，请重新设置" % args[0]
        if not os.path.isdir(dir_path):
            return "路径%s不是目录，请重新设置" % args[0]
        return ""

    @classmethod
    def get_key(cls) -> str:
        return "set_config_dir"

    def get_help(self) -> str:
        return '设置配置文件目录路径，支持 " %s <目录路径> " 设置，或 " %s ? " 查看详情' % (
            self.get_key(),
            self.get_key(),
        )

    def get_detail(self) -> str:
        return """
        设置配置文件目录路径，目录中可包含以下配置文件：

        1. 机房位置配置文件：LLD.xlsx
           - "灵衢L1网络对应关系" Sheet：包含列 服务器、机房名称、机柜编号、主机SN、L1名称、L1_IP、L1_SN
           - "灵衢L2网络对应关系" Sheet：包含列 设备名、机房名称、机柜编号、管理IP配置、SN

        2. 阈值覆盖配置文件：threshold_config.json
           - JSON对象：键为阈值名，值为 {"threshold": {要覆盖的阈值字段}}，未配置的阈值或字段保持默认值
           - 支持字段：low_value_alarm、high_value_alarm、low_value_warn、high_value_warn、normal_value_alarm、normal_value_warn等

        通过 " %s <目录路径> " 设置后，工具会自动扫描目录中的配置文件并加载
        """ % self.get_key()

    def add_arguments(self, parser):
        parser.add_argument(
            "action",
            metavar='actions',
            help="?(%s)=查看%s详细信息；目录路径=设置配置文件目录路径" % ("？", self.get_key()),
        )

    def run_task(self, *args) -> str:
        check_res = self.check_input_path(*args)
        if check_res:
            return check_res
        dir_path = convert_log_path(args[0])
        self.diag_ctx.read_from_dir(dir_path)
        return f"设置成功，配置目录：{dir_path}"


class SetConnConfigCliModel(DetailedCliModel):
    @staticmethod
    def is_support_param():
        return True

    @staticmethod
    def check_input_path(*args) -> str:
        if not args:
            return "地址为空，请重新设置"
        file_path = convert_log_path(args[0])
        if not file_path:
            return f"地址{args[0]}不存在，请重新设置"
        if not os.path.isfile(file_path):
            return f"地址{args[0]}不是文件，请重新设置"
        return ""

    @classmethod
    def get_key(cls) -> str:
        return "set_conn_config"

    def get_help(self) -> str:
        return f'设置连接文件地址，支持 " {self.get_key()} <文件地址> " 设置，或 " {self.get_key()} ? " 查看详情'

    def get_detail(self) -> str:
        return f"""
        配置文件内容结构样例
        ============== 样例开始 ==============

        [host]
        # port指定端口，不写默认为22，username指定用户名，password指定密码，private_key指定私钥文件
        1.1.1.1 port="22" username="root" private_key="~/.ssh/your_private_key"
        1.1.2.1 port="22" username="root" password="321"

        [bmc]
        1.1.1.2 username="Administrator" password="123"

        [switch]
        # 支持ip1-ip2 ip段方式填写(需保证账号密码相同)，通过step设置步长，如1.1.1.1-1.1.1.5 step=2 则得到1.1.1.1, 1.1.1.3, 1.1.1.5
        1.1.1.3-1.1.1.10 step=1 username="root" password="123"

        [config]
        # 支持设置全局的私钥文件
        private_key="~/.ssh/your_private_key"

        ============== 样例结束 ==============

        请在本机根据以上文件内容结构，编写需要远程连接的设备信息，保存到文件中。通过 " {self.get_key()} <文件地址> " 设置该文件后，工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动登录设备在线采集信息
        """

    def add_arguments(self, parser):
        parser.add_argument(
            "action", metavar='actions', help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置连接配置文件路径"
        )

    def run_task(self, *args) -> str:
        check_res = self.check_input_path(*args)
        if check_res:
            return check_res
        # 加密配置文件内容
        self.diag_ctx.encrypt_conn_config(args[0])
        # 加载配置
        res = self.diag_ctx.load_conn_config()
        if res:
            return f"设置地址失败，异常：{res}"
        return "设置成功，请尽快删除包含明文密码的配置文件"


class SetHostDumpLogDirCliModel(DetailedCliModel):
    @staticmethod
    def is_support_param():
        return True

    @classmethod
    def get_key(cls) -> str:
        return "set_host_dump_log"

    def get_help(self) -> str:
        return f'设置服务器导出日志目录，支持 " {self.get_key()} <目录> " 设置目录，或 " {self.get_key()} ? " 查看详情'

    def add_arguments(self, parser):
        parser.add_argument(
            "action",
            nargs='?',
            metavar='actions',
            help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置服务器导出日志目录",
        )

    def get_detail(self) -> str:
        return f"""
        设置服务器导出日志目录，支持以下几类脚本采集的日志:
        1. tool_log_collection_out_version_all_<version>.sh
        2. device_log_collect_<version>.sh
        3. link_down_collect_<version>.sh

        通过以上方式采集的日志压缩包，统一放到一个目录中，通过此命令 " {self.get_key()} <目录> " 设置目录，工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动解压分析日志信息
        """

    def run_task(self, *args) -> str:
        check_res = self.check_input_path(*args)
        if check_res:
            return check_res
        self.diag_ctx.dump_log_dir_config.host_dump_log_dir = args[0]
        return "设置成功"


class SetBmcDumpLogDirCliModel(DetailedCliModel):
    @staticmethod
    def is_support_param():
        return True

    @classmethod
    def get_key(cls) -> str:
        return "set_bmc_dump_log"

    def get_help(self) -> str:
        return f'设置BMC导出日志目录，支持 " {self.get_key()} <目录> " 设置目录，或 " {self.get_key()} ? " 查看详情'

    def add_arguments(self, parser):
        parser.add_argument(
            "action", nargs='?', metavar='actions', help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置BMC日志目录"
        )

    def get_detail(self) -> str:
        return f"""
        设置BMC导出日志目录，支持以下方式导出的日志tar.gz包
        1. 手动通过bmc网页 '一键收集' 按钮下载
        2. 通过命令 `ipmcget -d diaginfo` 采集的日志

        通过以上方式采集的日志压缩包，统一放到一个目录中，通过此命令 " {self.get_key()} <目录> " 设置目录，工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动解压分析日志信息
        """

    def run_task(self, *args) -> str:
        check_res = self.check_input_path(*args)
        if check_res:
            return check_res
        self.diag_ctx.dump_log_dir_config.bmc_dump_log_dir = args[0]
        return "设置成功"


class SetSwiDumpLogDirCliModel(DetailedCliModel):
    @staticmethod
    def is_support_param():
        return True

    @classmethod
    def get_key(cls) -> str:
        return "set_switch_dump_log"

    def get_help(self) -> str:
        return f'设置交换机命令回显/日志导出目录，支持 " {self.get_key()} <目录> " 设置目录，或 " {self.get_key()} ? " 查看详情'

    def add_arguments(self, parser):
        parser.add_argument(
            "action",
            nargs='?',
            metavar='actions',
            help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置交换机日志目录",
        )

    def get_detail(self) -> str:
        return f"""
        设置交换机命令回显/日志导出目录，支持以下方式导出的信息(当前仅支持华为交换机)
        1. 使用交换机 ' display diagnostic-information <filename>.txt ' 命令导出命令回显结果集(推荐，信息较全)
        2. 查询关键命令后直接复制shell回显页面，导出文本文件(必须执行display current-configuration获取交换机信息，否则工具无法匹配)
        3. 使用交换机 ' collect diagnostic information ' 命令导出的日志zip包
        将以上方式采集的文本文件统一放到一个目录中，通过此命令 " {self.get_key()} <目录> " 设置目录，工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动分析文本信息
        """

    def run_task(self, *args) -> str:
        check_res = self.check_input_path(*args)
        if check_res:
            return check_res
        self.diag_ctx.dump_log_dir_config.switch_dump_log_dir = args[0]
        return "设置成功"


class CollectBmcDumpInfoLog(CliModel):
    @classmethod
    def get_key(cls) -> str:
        return "collect_bmc_dump_info"

    def get_help(self) -> str:
        return "在线收集BMC dump info 日志"

    def run_task(self, *args) -> str:
        asyncio.run(CollectBmcLog(self.diag_ctx).main())
        return f"收集完成，请查看日志路径{CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR}"


class AutoCollectCliModel(DetailedCliModel):
    @classmethod
    def get_key(cls) -> str:
        return "auto_collect"

    def get_help(self) -> str:
        return "启动自动信息采集，支持离线、在线采集，适用于不同网络平面分批收集"

    # pylint: disable=useless-parent-delegation  # 父类是抽象方法，子类必须实现
    def get_detail(self) -> str:
        return super().get_detail()

    def run_task(self, *args) -> str:
        asyncio.run(AutoCollect(self.diag_ctx).main())
        return '收集完成，若完成全部收集请进行诊断/巡检'


class AutoInspection(DetailedCliModel):
    @classmethod
    def get_key(cls) -> str:
        return "auto_inspection"

    @staticmethod
    def is_support_param():
        return True

    def get_help(self) -> str:
        return "启动巡检结果诊断，适用于分批收集后统一诊断"

    def get_detail(self) -> str:
        # pylint: disable=useless-parent-delegation
        all_customer_types = "\n".join(customer.value for customer in list(Customer))
        return f"""
        使用 " auto_collect " 完成后，启动该命令进行巡检结果诊断。
        支持以下客户类型：
        {all_customer_types}
        使用 " {self.get_key()} <客户类型> " 启动巡检
        """

    def add_arguments(self, parser):
        parser.add_argument(
            "action",
            nargs='?',
            choices=['?', '？'] + [member.value for member in Customer],
            metavar='actions',
            help=f"?(？)=查看{self.get_key()}详细信息；客户类型=指定客户类型；无参数=采用默认客户类型",
        )

    def run_task(self, *args) -> str:
        if not args:
            _CONSOLE_LOGGER.info("未输入巡检类型，使用默认客户巡检")
            customer = Customer.DEFAULT
        else:
            customer = diag_enum.get_enum(Customer, "", args[0])
            if not customer:
                return f"{args[0]}为不支持的客户类型，请使用 ' {self.get_key()} ? ' 查看支持的客户类型"
        asyncio.run(Inspection(self.diag_ctx, customer).main())
        return "巡检完成"


class AutoDiagCliModel(DetailedCliModel):
    @classmethod
    def get_key(cls) -> str:
        return "auto_diag"

    def get_help(self) -> str:
        return "启动自动诊断，适用于分批收集后统一诊断"

    # pylint: disable=useless-parent-delegation  # 父类是抽象方法，子类必须实现
    def get_detail(self) -> str:
        return super().get_detail()

    def run_task(self, *args) -> str:
        try:
            asyncio.run(AutoDiagCluster(self.diag_ctx).main())
            return "诊断完成"
        except GenerateCsvPermissionErr as e:
            _CONSOLE_LOGGER.info(e)
            return f"生成报告失败，解除占用后，可使用 ' {self.get_key()} ' 重新生成报告。"


class AutoSingleDiagCliModel(DetailedCliModel):
    @staticmethod
    def is_support_param():
        return True

    @classmethod
    def max_param_count(cls) -> int:
        return 2

    @classmethod
    def get_key(cls) -> str:
        return "auto_single_diag"

    def get_help(self) -> str:
        return ('指定单条链路进行诊断，支持 " %s <ip> <id> " 指定设备IP和端口id，或 " %s ? " 查看详情') % (
            self.get_key(),
            self.get_key(),
        )

    def get_detail(self) -> str:
        return """
        指定单条链路进行诊断，IP自动识别设备类型（host/switch/bmc），端口id含义随设备类型不同：

        1. host：端口id为NPU ID（npu_id，取值0~7）
        2. switch：端口id为交换机端口名（如400GE1/0/1:10）
        3. bmc：端口id为关联主机的NPU ID（npu_id，取值0~7）

        执行全量诊断与根因分析，结果展示时仅保留该链路（本端或对端命中IP+端口id）的诊断结果：
        指定host端口id时，交换机侧对应该端口id的故障（携带对端信息）也会展示，反之亦然。

        使用 " %s <ip> <id> " 启动单链路诊断，示例：
        %s 1.1.1.1 5              # 诊断主机1.1.1.1的NPU端口5链路
        %s 1.1.1.3 400GE1/0/1:10  # 诊断交换机1.1.1.3的端口400GE1/0/1:10链路
        %s 1.1.1.2 5              # 诊断BMC 1.1.1.2关联主机NPU端口5链路

        执行前需确保缓存目录中存在结构化数据（通过 " %s " 或 " %s " 采集生成）
        """ % (
            self.get_key(),
            self.get_key(),
            self.get_key(),
            self.get_key(),
            AutoCollectCliModel.get_key(),
            AutoCollectDiagCliModel.get_key(),
        )

    def add_arguments(self, parser):
        parser.add_argument("ip", help="设备IP地址（host/switch/bmc）")
        parser.add_argument(
            "id", help="端口id：host/bmc为NPU端口号npu_id（0~7），switch为交换机端口名（如400GE1/0/1:10）"
        )

    def run_task(self, *args) -> str:
        if len(args) < 2:
            return (
                f"参数不足，用法：{self.get_key()} <ip> <id>，"
                f"ip为host/switch/bmc设备IP，id为端口id（host/bmc为npu_id 0~7，switch为交换机端口名），"
                f"使用 ' {self.get_key()} ? ' 查看详情"
            )
        ip, port = args[0], args[1]
        try:
            auto_single_diag = AutoSingleDiag(self.diag_ctx, ip, port)
            asyncio.run(auto_single_diag.main())
            if auto_single_diag.error_msg:
                return auto_single_diag.error_msg
            fault_count = len(self.diag_ctx.diag_result)
            return f"单链路诊断完成，共发现 {fault_count} 条故障，诊断结果仅展示该链路"
        except GenerateCsvPermissionErr as e:
            _CONSOLE_LOGGER.info(e)
            retry_cmd = f"{self.get_key()} {ip} {port}"
            return f"生成报告失败，解除占用后，可使用 ' {retry_cmd} ' 重新生成报告。"


class AutoCollectDiagCliModel(CliModel):
    @classmethod
    def get_key(cls) -> str:
        return "auto_collect_diag"

    def get_help(self) -> str:
        return "启动一键式自动收集（在线设备采集或离线日志收集）诊断"

    def run_task(self, *args) -> str:
        try:
            asyncio.run(AutoCollect(self.diag_ctx).main())
            asyncio.run(AutoDiagCluster(self.diag_ctx).main())
            return "诊断完成"
        except GenerateCsvPermissionErr as e:
            _CONSOLE_LOGGER.info(e)
            return f"生成报告失败，解除占用后，可使用 ' {self.get_key()} ' 重新生成报告。"


class ClearCacheCliModel(CliModel):
    @classmethod
    def get_key(cls) -> str:
        return "clear_cache"

    def get_help(self) -> str:
        return "清理缓存，请在执行新诊断任务前务必执行！避免干扰诊断结果（若清理未生效请用管理员模式打开工具）"

    def run_task(self, *args) -> str:
        try:
            if os.path.exists(CommonPath.TOOL_HOME_CACHE_DIR):
                shutil.rmtree(CommonPath.TOOL_HOME_CACHE_DIR)
            if os.path.isfile(CommonPath.ENCRYPTED_CONN_CONFIG_PATH):
                os.remove(CommonPath.ENCRYPTED_CONN_CONFIG_PATH)
        except Exception as e:
            return f"清理{CommonPath.COLLECT_CACHE}异常：{e}"
        return "清理完成"


_LOCAL_VARS = dict(vars().items())


def build_cli_ctx(diag_ctx: DiagCtx) -> CliCtx:
    cli_models = []
    cli_ctx = CliCtx()
    for _, cls in _LOCAL_VARS.items():
        if isinstance(cls, type) and issubclass(cls, CliModel) and not inspect.isabstract(cls):
            cli_models.append(cls(diag_ctx, cli_ctx))
    cli_ctx.update_cli_models(cli_models)
    return cli_ctx
