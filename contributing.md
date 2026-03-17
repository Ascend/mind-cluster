[TOC]

# 开发MindCluster
请先按组件分类阅读文档：
* [ascend-common](https://gitcode.com/Ascend/mind-cluster/blob/master/component/ascend-common/README.md)
* [ascend-device-plugin](https://gitcode.com/Ascend/mind-cluster/blob/master/component/ascend-device-plugin/README.md)
* [ascend-docker-runtime](https://gitcode.com/Ascend/mind-cluster/blob/master/component/ascend-docker-runtime/README.md)
* [ascend-for-volcano](https://gitcode.com/Ascend/mind-cluster/blob/master/component/ascend-for-volcano/README.md)
* [ascend-operator](https://gitcode.com/Ascend/mind-cluster/blob/master/component/ascend-operator/README.md)
* [clusterd](https://gitcode.com/Ascend/mind-cluster/blob/master/component/clusterd/README.md)
* [container-manager](https://gitcode.com/Ascend/mind-cluster/blob/master/component/container-manager/README.md)
* [mindcluster-tools](https://gitcode.com/Ascend/mind-cluster/blob/master/component/mindcluster-tools/README.MD)
* [noded](https://gitcode.com/Ascend/mind-cluster/blob/master/component/noded/README.md)
* [npu-exporter](https://gitcode.com/Ascend/mind-cluster/blob/master/component/npu-exporter/README.md)
* [taskd](https://gitcode.com/Ascend/mind-cluster/blob/master/component/taskd/README.md)

# 代码提交规范
## Commit 消息格式
所有提交必须遵循以下格式：
```
<type>【component】: <subject>

<body>
```
## Type（类型）
feat: 新功能
fix: Bug修复
docs: 文档更新
style: 代码格式（不影响功能，如空格、格式化等）
refactor: 重构（既不是 Bug 修复也不是新功能）
perf: 性能优化
test: 测试相关
chore: 构建/工具变更（如依赖更新、构建配置等）
ci: CI/CD 相关变更
## Component（组件）
指定提交涉及的组件范围，例如：
- `clusterD`
- `device-plugin`
## Subject（主题）
- 使用祈使句，首字母小写
- 不超过50个字符
- 不以句号结尾
- 描述"做了什么"而不是"做了什么改动"
## Body（正文，可选）
- 详细描述变更的原因和方式
- 说明与之前行为的对比
- 可以多行，每行不超过 72 个字符

# PullRequest
## PR 创建流程
1. **创建特性分支**
   ```bash
   git checkout -b feature/your-feature-name
   # 或
   git checkout -b fix/issue-number
   ```
2. **进行开发**
    - 编写代码
    - 添加测试
      - **Go测试**：建议使用convey框架编写测试用例
        ```go
        package utils
        
        import (
            "testing"
            
            "github.com/smartystreets/goconvey/convey"
            "k8s.io/apimachinery/pkg/apis/meta/v1"
        )
        
        func TestIsMindIEEPJob(t *testing.T) {
            convey.Convey("isMindIEEPJob", t, func() {
                job := &v1.AscendJob{}
                convey.Convey("job nil will return false", func() {
                    res := IsMindIEEPJob(nil)
                    convey.So(res, convey.ShouldBeFalse)
                })
                convey.Convey("label nil will return false", func() {
                    res := IsMindIEEPJob(job)
                    convey.So(res, convey.ShouldBeFalse)
                })
            })
        }
        ```
      - **Python测试**：建议使用unittest框架编写测试用例
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
      - **C/C++测试**：建议使用Google Test框架编写测试用例
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
    - 更新文档
    - 确保代码通过本地测试
3. **提交代码**
   ```bash
   git add .
   git commit -m "[feat] add new feature"
   ```
4. **推送到 Fork 仓库**
   ```bash
   git push origin feature/your-feature-name
   ```
5. **创建 Pull Request**
    - 访问gitcode仓库页面
    - 点击"Pull Request"或"合并请求"
    - 填写PR描述（见PR创建页面模板）

## PR最佳实践
1. **保持PR小规模**
    - 一次PR只解决一个问题
    - 便于评审和理解
    - 提高合并效率
    - 建议：单个PR不超过1000行（含测试）代码变更
2. **及时更新**
    - 定期同步上游主分支
    - 及时响应评审意见
    - 保持 PR 活跃
3. **清晰描述**
    - 详细描述变更原因和方式
    - 提供测试方法
    - 添加截图或示例（如适用）

## PR评审与合入规则
### 评审要求
1. **评审人员要求**
    - 评审人员必须熟悉相关代码领域
    - 评审人员不能是PR作者本人
2. **评审检查项**
    - ✅ 代码质量和风格
    - ✅ 功能正确性
    - ✅ 测试覆盖率（分支60%，行80%）
    - ✅ 文档完整性
    - ✅ 性能影响
    - ✅ 安全性
    - ✅ 向后兼容性
3. **CI 检查要求**
    - ✅ 所有 CI 检查必须通过
4. **无 Block 评论**
    - PR不能有任何未解决问题
### 合入规则
1. **Squash and Merge**
    - 将 PR 的所有提交合并为一个提交
    - 保持主分支历史清晰
    - 提交消息使用PR标题
2. **必须满足的条件**
    - ✅ 至少2位Maintainer或Committer的/lgtm，和1个/approve
3. **禁止的操作**
    - ❌ 禁止 Force Push 到主分支
    - ❌ 禁止合并自己的 PR（必须有他人评审）
### 合并权限
- **Maintainer**：可以合并任何PR
- **Committer**：可以合并任何PR
- **Contributor**：无合并权限，需要等待Maintainer或Committer合并

# CI说明
CI检查项目有：
* 执行Shell：CI内部调用。
* 执行Shell：CI内部调用。
* Build_arm：构建集群管理组件二进制包
* Build_x86：构建集群管理组件二进制包
* build_mindio_arm: 构建mindio软件包
* build_mindio_x86: 构建mindio软件包
* code_check：编码风格、规范、安全检查。
* anti_poison：病毒扫描。
* sca：开源合规检查。
* UT_go：go单元测试
* UT_cpp：cpp单元测试
* UT_python：python单元测试

任意一项失败可以通过详情链接查看具体问题。如果是CI自身故障，请[联系committer](https://gitcode.com/Ascend/community/blob/master/MindCluster/sigs/MindCluster/sig-info.yaml)，或通过评论“rebuild”尝试重新构建。

# Special Interest Group
## 工作目标和范围
1. 技术聚焦
   围绕基于NPU的集群全流程运行，提供集群作业调度、运维监测、故障恢复等功能进行深入研究，推动技术发展，解决实际问题。
2. 促进协作
   通过组织会议、技术分享等方式，促进成员之间的协作和知识共享，提升技术水平。
3. 最佳实践
   在技术实现、接口设计、开发流程等方面推动最佳实践，降低协作成本，提升系统兼容性和可维护性。
4. 社区建设
   通过代码贡献、技术分享等方式，培养技术人才，推动社区生态建设。

## 例会
* 周期：每1个月举行一次例会，可通过[Ascend开源社区](https://meeting.ascend.osinfra.cn/)搜索、查看sig-MindCluster的会议链接。
* 申报议题：通过[sig-MindCluster Etherpad链接](https://etherpad.ascend.osinfra.cn/p/sig-MindCluster)进入共享文档，编辑申报议题。
* 参会人员：maintainer、committer、contributor等核心成员，其他对本SIG感兴趣的人员。
* 会议内容：讨论遗留问题和进展；当期申报的议题；需求评审、任务和优先级；需求规划和进展（roadmap）；新晋maintainer、committer准入评审。
* 会议归档：会议纪要位于[sig-MindCluster Etherpad链接](https://etherpad.ascend.osinfra.cn/p/sig-MindCluster)。

## 成员列表
[SIG成员列表](https://gitcode.com/Ascend/community/blob/master/MindCluster/sigs/MindCluster/sig-info.yaml)。
