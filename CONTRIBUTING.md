# 为Ascend MindCluster贡献

  感谢您考虑为 Ascend MindCluster 做出贡献！我们欢迎任何形式的贡献，包括错误修复、功能增强、文档改进等，甚至只是反馈。无论您是经验丰富的开发者还是第一次参与开源项目，您的帮助都是非常宝贵的。

  您可以通过多种方式支持本项目：

  - 通过[MindCluster社区任务池](https://gitcode.com/Ascend/mind-cluster/issues/154)参与贡献。
  - 审查Pull Request并协助其他贡献者。
  - 传播项目：在博客文章、社交媒体上分享MindCluster，或给仓库点个⭐。

  参与贡献前，请先签署开放项目贡献者许可协议（[CLA](https://link.gitcode.com/?target=https%3A%2F%2Fclasign.osinfra.cn%2Fsign%2Fgitee_ascend-1611222220829317930&from=https%3A%2F%2Fgitcode.com%2FAscend%2FIndexSDK%2Fblob%2Fmaster%2FCONTRIBUTING.md&lang=zh&theme=white)），提前了解[《社区行为规范》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/code-of-conduct.md)。

## 贡献方式

### Pull Request

  提交PR前，请先了解[《Pull Request (PR) 提交流程指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/pr-guide.md) 、[PR最佳实践](#pr最佳实践)和[单元测试最佳实践](#单元测试最佳实践)，掌握从Fork到提交、从代码审查到合并的完整PR流程，包括CI检查、标签要求和合并规范。

### Issue

  提交Issue前，请先了解[《Issue 创建与处理指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/issue-guide.md) ，学习如何有效创建、分类和管理Issue，包括问题报告、功能请求的规范格式。

### SIG会议

  每1个月举行一次例会，可提前通过[sig-MindCluster Etherpad链接](https://link.gitcode.com/?target=https%3A%2F%2Fetherpad.ascend.osinfra.cn%2Fp%2Fsig-MindCluster&from=https%3A%2F%2Fgitcode.com%2FAscend%2Fmind-cluster%2Fblob%2Fmaster%2Fcontributing.md&lang=zh&theme=white)进入共享文档 ，编辑申报议题，在[MindCluster SIG会议日历](https://meeting.ascend.osinfra.cn/?sig=sig-MindCluster)中找到对应会议链接并按时参会。

  [SIG成员列表](https://gitcode.com/Ascend/community/blob/master/MindCluster/sigs/MindCluster/sig-info.yaml)。

## PR最佳实践

  1. **Fork仓库**：在GitCode平台代码仓库右上角点击"Fork"按钮，Fork一份源代码到个人仓。

  2. **克隆到本地**：

     将Fork到个人仓的代码克隆到本地进行代码开发。

     ```bash
     git clone https://gitcode.com/<your-username>/mind-cluster.git
     cd mind-cluster
     ```

  3. **创建特性分支**

     ```bash
     git checkout -b feature/your-feature-name
     # 或
     git checkout -b fix/issue-number
     ```

  4. **代码开发**

     - 编写代码，质量符合[开发规范](#dev-rule)和[安全编程指导](#sec-guide)。

     - 如涉及文档修改的，需同步完成文档更新。
     - 确保代码通过本地测试，可以参考[单元测试最佳实践](#单元测试最佳实践)。

  5. **开发构建验证**

     - 非ascend-faultdiag组件请参考[编译指南](build/README.md)部署编译环境并完成组件编译构建。构建完成后，参考[安装指南](./docs/zh/scheduling/03_installation_guide/menu_installation_guide.md)安装并验证组件（安装前请先阅读[安装前必读](./docs/zh/scheduling/03_installation_guide/00_before_you_start.md)和[环境依赖](./docs/zh/scheduling/03_installation_guide/01_environment_dependencies.md)）。
     - Ascend Faultdiag可参考日志诊断工具[安装](./docs/zh/faultdiag/ascend-faultdiag/04_installation_guide/01_installation.md)或链路诊断工具[安装](./docs/zh/faultdiag/ascend-faultdiag-toolkit/04_installation_guide/01_installation.md)，完成组件编译，安装和功能验证。

  6. **执行pre-commit检查**

     本地提交代码前请先执行pre-commit检查，检查指导参见[pre-commit本地运行指南](https://gitcode.com/Ascend/community/blob/master/docs/contributor/pre-commit-guide.md)。

  7. **提交Pull Request**

     - 保持PR小规模，一次PR只解决一个问题，单个PR不超过1000行（含测试）代码变更。
     - 及时更新，定期同步上游主分支，及时响应评审意见。
     - 描述清晰，详细描述变更原因和方式，提供测试方法，添加截图或示例。

  8. **社区评审**

     如果涉及patch、头文件宏、API接口等更新，需提交社区在SIG例会进行评审，社区定期例会与活动参见[会议日历](https://meeting.ascend.osinfra.cn/?sig=sig-MindCluster)。

## 单元测试最佳实践

  1. 单元测试开发

     - **Go测试**：建议使用convey框架编写测试用例。

       ```go
       package common

       import (
           "testing"
           "github.com/smartystreets/goconvey/convey"
       )

       // TestAtomicBool for test AtomicBool
       func TestAtomicBool(t *testing.T) {
           convey.Convey("test AtomicBool", t, func() {
               flag := NewAtomicBool(false)
               ret := flag.Load()
               convey.So(ret, convey.ShouldBeFalse)
               flag.Store(true)
               ret = flag.Load()
               convey.So(ret, convey.ShouldBeTrue)
           })
       }
       ```

     - **Python测试**：建议使用unittest框架编写测试用例。

       ```python
       #!/usr/bin/env python3

       import unittest
       from unittest.mock import patch, MagicMock

       class TestTaskdManagerAPI(unittest.TestCase):
           @patch('taskd.api.taskd_manager_api.Manager')
           def test_init_taskd_manager_success(self, mock_manager):
               # mock the init_taskd_manager method
               mock_manager_instance = MagicMock()
               mock_manager_instance.init_taskd_manager.return_value = True
               mock_manager.return_value = mock_manager_instance

               config = {}
               result = init_taskd_manager(config)

               # verify method calls
               mock_manager.assert_called_once()
               mock_manager_instance.init_taskd_manager.assert_called_once_with(config)
               # verify return value
               self.assertEqual(result, True)
       ```

     - **C/C++测试**：建议使用Google Test框架编写测试用例。

       ```cpp
       #include <gtest/gtest.h>
       #include "file_utils.h"

       using namespace ock::ttp;

       class FileUtilsTest : public ::testing::Test {
       protected:
           void SetUp() override
           {
               // 设置测试环境
               testFile = "/tmp/test_file.txt";
               std::ofstream file(testFile);
               file << "Test content";
               file.close();
           }

           void TearDown() override
           {
               // 清理测试环境
               remove(testFile.c_str());
           }

           std::string testFile;
       };

       TEST_F(FileUtilsTest, CheckFileExists_WhenFileExists_ReturnsTrue)
       {
           ASSERT_TRUE(FileUtils::CheckFileExists(testFile));
       }

       TEST_F(FileUtilsTest, CheckFileExists_WhenFileNotExists_ReturnsFalse)
       {
           std::string nonExistentFile = "/tmp/non_existent.txt";
           ASSERT_FALSE(FileUtils::CheckFileExists(nonExistentFile));
       }
       ```

  2. 执行用例

        - 非Ascend Faultdiag组件，编译或执行用例前，按以下步骤操作：

          1. 在对应组件目录（即包含 `go.mod` 的目录）下执行以下命令，命令会自动补全和校正 `go.mod` 与 `go.sum`，避免因依赖缺失或不一致导致编译或测试失败。

             ```bash
             go mod tidy
             ```

          2. 在对应组件的 `build` 目录下执行测试脚本：

             ```bash
             cd build
             ./test.sh
             ```

        - Ascend Faultdiag执行UT用例验证，需确保环境中Python >= 3.8，且python3，pip3命令可用。在项目根目录下执行以下命令：

          ```bash
          pip3 install -r ./component/ascend-faultdiag/src/requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple
          pip3 install pytest pytest-cov pytest-html -i https://pypi.tuna.tsinghua.edu.cn/simple
          bash component/ascend-faultdiag/test/run_dt.sh
          ```

## 分支/Tag命名规则

  - 从2026年8月开始，新创建的分支、Tag命名规则如下：

    | 分支类型  | 分支名规则            | 示例                              | 说明         | tag名规则                     | tag示例                              |
    | --------- | --------------------- | --------------------------------- | ------------ | ----------------------------- | ------------------------------------ |
    | 主干&开发 | master                | -                                 | -            | -                             | -                                    |
    | release   | release/<版本号>      | release/v26.1.0                   | 正式版本     | <版本号>[-beta.<序号>]        | v26.1.0，v26.1.0-beta.1              |
    | poc       | poc/<基线分支>/<描述> | poc/release-v26.1.0/auth-redesign | 后续合入主干 | poc/<基线分支>/<描述>-v<序号> | poc/release-v26.1.0/auth-redesign-v1 |

## 参考

  - 开发规范<a id="dev-rule"></a>
    - [《Ascend C++ 编码风格指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-cpp-coding-style-guide.md)
    - [《Ascend Python 编码风格指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-python-coding-style-guide.md)
    - [《Ascend Go 编码风格指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-go-coding-style-guide.md)
  - 安全编程指导<a id="sec-guide"></a>
    - [《Ascend C++ 安全编程指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-cpp-secure-coding-guide.md)
    - [《Ascend Python 安全编程指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-python-secure-coding-guide.md)
    - [《Ascend Go 安全编程指南》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-go-secure-coding-guide.md)
  - [《Ascend安全编译选项指南(C&C++)》](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-secure-compile-guide.md)

  - 更多社区相关规范，请访问[Ascend社区community](https://gitcode.com/Ascend/community)
