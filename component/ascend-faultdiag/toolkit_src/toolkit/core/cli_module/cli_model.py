import asyncio
import inspect
import os.path
import shutil

from toolkit.core.cli_module.base import CliModel, DetailedCliModel, CliCtx
from toolkit.core.common import diag_enum
from toolkit.core.common.diag_enum import Customer
from toolkit.core.common.errors import GenerateCsvPermissionErr
from toolkit.core.common.path import CommonPath
from toolkit.core.context.diag_ctx import DiagCtx
from toolkit.examples.auto_diag.auto_collect import AutoCollect
from toolkit.examples.auto_diag.auto_diag import AutoDiagCluster
from toolkit.examples.auto_diag.collect_bmc_log import CollectBmcLog
from toolkit.examples.inspection.inspection import Inspection
from toolkit.utils import logger

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER


class HelpCliModel(CliModel):
    _SPACE_SIZE = 6

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

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
    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "exit"

    def get_help(self) -> str:
        return "退出程序"

    def run_task(self, *args) -> str:
        self.cli_ctx.is_running = False
        return "再见!"


class ClearCliModel(CliModel):
    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "clear"

    def get_help(self) -> str:
        return "清屏"

    def run_task(self, *args) -> str:
        os.system('cls' if os.name == 'nt' else 'clear')
        return ""


class AboutCliModel(CliModel):
    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "about"

    def get_help(self) -> str:
        return "查看关于诊断工具"

    def run_task(self, *args) -> str:
        return f"""
        MindCluster ascend-faultdiag-toolkit诊断工具版本: {self.diag_ctx.tool_config.version}
        """


class GuideCliModel(CliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "guide"

    def get_help(self) -> str:
        return "获取向导信息"

    def run_task(self, *args) -> str:
        return f"""
        一. 采集内容准备
        请根据故障设备自行按需选择要采集的设备信息或需要导入的日志, 可以不导入全量设备信息或日志. 按需设置以下在线或离线采集分析的任意地址.
        
        1. 在线采集准备
        若需要在线采集设备信息, 请使用 " {SetConnConfigCliModel.get_key()} " 命令设置设备信息, 具体配置可使用" {SetConnConfigCliModel.get_key()} ? "查看详情
        
        2. 离线日志解析准备
        2.1 设置服务器日志目录地址
        请使用 " {SetHostDumpLogDirCliModel.get_key()} " 命令设置设备信息, 具体配置可使用" {SetHostDumpLogDirCliModel.get_key()} ? "查看详情
        
        2.2 设置BMC日志目录地址
        请使用 " {SetBmcDumpLogDirCliModel.get_key()} " 命令设置设备信息, 具体配置可使用" {SetBmcDumpLogDirCliModel.get_key()} ? "查看详情
        
        2.2 设置交换机回显文本目录地址
        请使用 " {SetSwiDumpLogDirCliModel.get_key()} " 命令设置设备信息, 具体配置可使用"{SetSwiDumpLogDirCliModel.get_key()} ? "查看详情
        
        3. 默认读取路径
        当未手动设置以上文件或目录时,工具会自动读取执行路径下的以下默认文件或目录
        连接配置: conn.ini
        BMC日志目录: bmc_dump_log
        Host日志目录: host_dump_log
        交换机日志目录: switch_dump_log        
        
        二. 启动采集/分析 & 诊断
        执行 " {AutoCollectDiagCliModel.get_key()} " 启动在线采集/离线分析并诊断
        
        三. 清理缓存
        本工具支持分批采集统一诊断, 所以会单次诊断完后会留有缓存, 若已完成诊断任务, 请使用 " {ClearCacheCliModel.get_key()} " 清理缓存(若无法有效清理, 请使用管理员模式打开工具), 避免影响下次诊断结果
 
        总结: 
        1. 先用 " {SetConnConfigCliModel.get_key()} " 设置要访问的设备ip配置文件或用 " {SetBmcDumpLogDirCliModel.get_key()} ", " {SetSwiDumpLogDirCliModel.get_key()} ", " {SetSwiDumpLogDirCliModel.get_key()} "设置离线日志目录, 或直接将日志放到默认目录下
        2. 以上至少有一项设置存在即可使用 " {AutoCollectDiagCliModel.get_key()} " 采集/分析并诊断输出报告        
        """


class SetConnConfigCliModel(DetailedCliModel):
    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "set_conn_config"

    def get_help(self) -> str:
        return f'设置连接文件地址, 支持 " {self.get_key()} <文件地址> " 设置, 或 " {self.get_key()} ? " 查看详情'

    def get_detail(self) -> str:
        return f"""
        配置文件内容结构样例
        ============== 样例开始 ============== 
        
        [host]
        # port指定端口,不写默认为22, username指定用户名, password指定密码, private_key指定私钥文件
        1.1.1.1 port="22" username="root" private_key="~/.shh/your_private_key"
        1.1.2.1 port="22" username="root" password="321" 
        
        [bmc]
        1.1.1.2 username="Administrator" password="123"
        
        [switch]
        # 支持ip1-ip2 ip段方式填写(需保证账号密码相同), 通过step设置步长, 如1.1.1.1-1.1.1.5 step=2 则得到1.1.1.1, 1.1.1.3, 1.1.1.5
        1.1.1.3-1.1.1.10 step=1 username="root" password="123"
        
        [config]
        # 支持设置全局的私钥文件
        private_key="~/.shh/your_private_key"
        
        ============== 样例结束 ==============
        
        请在本机根据以上文件内容结构, 编写需要远程连接的设备信息, 保存到文件中. 通过 " {self.get_key()} <文件地址> " 设置该文件后, 工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动登录设备在线采集信息
        """

    def add_arguments(self, parser):
        parser.add_argument("action", metavar='actions',
                            help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置连接配置文件路径")

    def run_task(self, *args) -> str:
        if not args:
            return "地址为空, 请重新设置"
        if not os.path.exists(args[0]):
            return f"地址{args[0]}不存在, 请重新设置"
        # 加密配置文件内容
        self.diag_ctx.encrypt_conn_config(args[0])
        # 加载配置
        res = self.diag_ctx.load_conn_config()
        if res:
            return f"设置地址失败, 异常: {res}"
        return "设置成功, 请尽快删除包含明文密码的配置文件"


class SetHostDumpLogDirCliModel(DetailedCliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "set_host_dump_log"

    def get_help(self) -> str:
        return f'设置服务器导出日志目录, 支持 " {self.get_key()} <目录> " 设置目录, 或 " {self.get_key()} ? " 查看详情'

    def add_arguments(self, parser):
        parser.add_argument("action", nargs='?', metavar='actions',
                            help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置服务器导出日志目录")

    def get_detail(self) -> str:
        return f"""
        设置服务器导出日志目录, 支持以下几类脚本采集的日志:
        1. A3device日志一键采集脚本<version>.sh
        2. link_down_collect_<version>.sh
        3. tool_log_collection_out_version_all_<version>.sh (以上脚本获取请联系昇腾维护, 或@wang-ruiju)
                
        通过以上方式采集的日志压缩包, 统一放到一个目录中, 通过此命令 " {self.get_key()} <目录> " 设置目录, 工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动解压分析日志信息
        """

    def run_task(self, *args) -> str:
        if not args:
            return "地址为空, 请重新设置"
        if not os.path.exists(args[0]):
            return f"地址{args[0]}不存在, 请重新设置"
        if not os.path.isdir(args[0]):
            return f"地址{args[0]}非文件夹, 请重新设置"
        self.diag_ctx.dump_log_dir_config.host_dump_log_dir = args[0]
        return "设置成功"


class SetBmcDumpLogDirCliModel(DetailedCliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "set_bmc_dump_log"

    def get_help(self) -> str:
        return f'设置BMC导出日志目录, 支持 " {self.get_key()} <目录> " 设置目录, 或 " {self.get_key()} ? " 查看详情'

    def add_arguments(self, parser):
        parser.add_argument("action", nargs='?', metavar='actions',
                            help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置MBC日志目录")

    def get_detail(self) -> str:
        return f"""
        设置BMC导出日志目录, 支持以下方式导出的日志tar.gz包
        1. 手动通过bmc网页 '一键收集' 按钮下载
        2. 通过命令 `ipmcget -d diaginfo` 采集的日志
        
        通过以上方式采集的日志压缩包, 统一放到一个目录中, 通过此命令 " {self.get_key()} <目录> " 设置目录, 工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动解压分析日志信息
        """

    def run_task(self, *args) -> str:
        if not args:
            return "地址为空, 请重新设置"
        if not os.path.exists(args[0]):
            return f"地址{args[0]}不存在, 请重新设置"
        if not os.path.isdir(args[0]):
            return f"地址{args[0]}非文件夹, 请重新设置"
        self.diag_ctx.dump_log_dir_config.bmc_dump_log_dir = args[0]
        return "设置成功"


class SetSwiDumpLogDirCliModel(DetailedCliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "set_switch_dump_log"

    def get_help(self) -> str:
        return f'设置交换机命令回显导出目录, 支持 " {self.get_key()} <目录> " 设置目录, 或 " {self.get_key()} ? " 查看详情'

    def add_arguments(self, parser):
        parser.add_argument("action", nargs='?', metavar='actions',
                            help=f"?(？)=查看{self.get_key()}详细信息；文件路径=设置交换机日志目录")

    def get_detail(self) -> str:
        return f"""
        设置交换机命令回显/日志导出目录, 支持以下方式导出的信息(当前仅支持华为交换机)
        1. 使用交换机 ' display diagnostic-information <filename> ' 命令导出命令回显结果集(推荐, 信息较全)
        2. 查询关键命令后直接复制shell回显页面, 导出文本文件(必须执行display current-configuration获取交换机信息, 否则工具无法匹配)
        3. 使用交换机 ' collect diagnostic-information ' 命令导出的日志zip包 
        将以上方式采集的文本文件统一放到一个目录中, 通过此命令 " {self.get_key()} <目录> " 设置目录, 工具会在 " {AutoCollectDiagCliModel.get_key()} " 命令下自动分析文本信息
        """

    def run_task(self, *args) -> str:
        if not args:
            return "地址为空, 请重新设置"
        if not os.path.exists(args[0]):
            return f"地址{args[0]}不存在, 请重新设置"
        if not os.path.isdir(args[0]):
            return f"地址{args[0]}非文件夹, 请重新设置"
        self.diag_ctx.dump_log_dir_config.switch_dump_log_dir = args[0]
        return "设置成功"


class CollectBmcDumpInfoLog(CliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "collect_bmc_dump_info"

    def get_help(self) -> str:
        return "在线收集BMC dump info日志"

    def run_task(self, *args) -> str:
        asyncio.run(CollectBmcLog(self.diag_ctx).main())
        return f"收集完成, 日志位于{CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR}"


class AutoCollectCliModel(DetailedCliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "auto_collect"

    def get_help(self) -> str:
        return "启动自动信息采集, 支持离线、在线采集, 适用于不同网络平面分批收集"

    def get_detail(self) -> str:
        return super().get_detail()

    def run_task(self, *args) -> str:
        asyncio.run(AutoCollect(self.diag_ctx).main())
        return f'收集完成, 若完成全部收集请使用 " {AutoDiagCliModel.get_key()} " 进行诊断'


class AutoInspection(DetailedCliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "auto_inspection"

    def get_help(self) -> str:
        return "启动巡检结果诊断, 适用于分批收集后统一诊断"

    def get_detail(self) -> str:
        all_customer_types = "\n".join(customer.value for customer in list(Customer))
        return f"""
        使用 " auto_collect " 完成后, 启动该命令进行巡检结果诊断. 
        支持以下客户类型: 
        {all_customer_types}
        使用 " {self.get_key()} <客户类型> " 启动诊断
        """

    def add_arguments(self, parser):
        parser.add_argument("action", nargs='?', choices=['?', '？'] + [member.value for member in Customer],
                            metavar='actions',
                            help=f"?(？)=查看{self.get_key()}详细信息；客户类型=指定客户类型；无参数=采用默认客户类型")

    def run_task(self, *args) -> str:
        if not args:
            _CONSOLE_LOGGER.info("未输入巡检类型, 默认使用mayi客户巡检")
            customer = Customer.Mayi
        else:
            customer = diag_enum.get_enum(Customer, "", args[0])
            if not customer:
                return f"{args[0]}为不支持的客户类型, 请使用 ' {self.get_key()} ? ' 查看支持的客户类型"
        asyncio.run(Inspection(self.diag_ctx, customer).main())
        return "诊断完成"


class AutoDiagCliModel(DetailedCliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "auto_diag"

    def get_help(self) -> str:
        return "启动自动诊断, 适用于分批收集后统一诊断"

    def get_detail(self) -> str:
        return super().get_detail()

    def run_task(self, *args) -> str:
        try:
            asyncio.run(AutoDiagCluster(self.diag_ctx).main())
            return "诊断完成"
        except GenerateCsvPermissionErr as e:
            _CONSOLE_LOGGER.info(e)
            return "生成csv失败, 解除占用后, 可使用 ' auto_diag ' 重新生成报告."


class AutoCollectDiagCliModel(CliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "auto_collect_diag"

    def get_help(self) -> str:
        return "启动一键式自动收集(在线设备采集或离线日志收集)诊断"

    def run_task(self, *args) -> str:
        try:
            asyncio.run(AutoCollect(self.diag_ctx).main())
            asyncio.run(AutoDiagCluster(self.diag_ctx).main())
            return "诊断完成"
        except GenerateCsvPermissionErr as e:
            _CONSOLE_LOGGER.info(e)
            return "生成csv失败, 解除占用后, 可使用 ' auto_diag ' 重新生成报告."


class ClearCacheCliModel(CliModel):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        super().__init__(diag_ctx, cli_ctx)

    @classmethod
    def get_key(cls) -> str:
        return "clear_cache"

    def get_help(self) -> str:
        return "清理缓存, 请在执行新诊断任务前务必执行! 避免干扰诊断结果(若清理未生效请用管理员模式打开工具)"

    def run_task(self, *args) -> str:
        try:
            if os.path.exists(CommonPath.COLLECT_CACHE):
                shutil.rmtree(CommonPath.COLLECT_CACHE)
            if os.path.isfile(CommonPath.ENCRYPTED_CONN_CONFIG_PATH):
                os.remove(CommonPath.ENCRYPTED_CONN_CONFIG_PATH)
        except Exception as e:
            return f"清理{CommonPath.COLLECT_CACHE}异常: {e}"
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
