#include <gtest/gtest.h>
#include <fstream>
#include <sys/stat.h>
#include <unistd.h>
#include <dirent.h>
#include <cstdlib>
#include <sstream>
#include "file_utils.h"

using namespace ock::ttp;

class FileUtilsTest : public ::testing::Test {
protected:
    void SetUp() override
    {
        std::string pid_str = std::to_string(getpid());
        testDir = "/tmp/FileUtilsTest_" + pid_str;
        testFile = testDir + "/test_file.txt";
        testLink = testDir + "/test_link.txt";
        testSubDir = testDir + "/subdir";
        restrictedFile = testDir + "/restricted.txt";

        mkdir(testDir.c_str(), 0755);
        mkdir(testSubDir.c_str(), 0755);

        std::ofstream file(testFile);
        file << "Test content";
        file.close();

        symlink(testFile.c_str(), testLink.c_str());

        std::ofstream restricted(restrictedFile);
        restricted << "Restricted content";
        restricted.close();
        chmod(restrictedFile.c_str(), 0400);
    }
    
    void TearDown() override
    {
        system(("rm -rf " + testDir).c_str());
    }
    
    std::string testDir;
    std::string testFile;
    std::string testLink;
    std::string testSubDir;
    std::string restrictedFile;
};

TEST_F(FileUtilsTest, IntegrationTest_FileValidationWorkflow)
{
    std::string errMsg;

    ASSERT_TRUE(FileUtils::CheckFileExists(testFile));

    ASSERT_TRUE(FileUtils::RegularFilePath(testFile, testDir, errMsg));

    ASSERT_TRUE(FileUtils::IsFileValid(testFile, errMsg, true, false));

    ASSERT_TRUE(FileUtils::CheckOwner(testFile, errMsg));

    chmod(testFile.c_str(), 0644);
    ASSERT_TRUE(FileUtils::CheckPermission(testFile, 0644, false, errMsg));
}

// test CheckFileExists
TEST_F(FileUtilsTest, CheckFileExists_WhenFileExists_ReturnsTrue)
{
    ASSERT_TRUE(FileUtils::CheckFileExists(testFile));
}

TEST_F(FileUtilsTest, CheckFileExists_WhenFileNotExists_ReturnsFalse)
{
    std::string nonExistentFile = testDir + "/non_existent.txt";
    ASSERT_FALSE(FileUtils::CheckFileExists(nonExistentFile));
}

TEST_F(FileUtilsTest, CheckFileExists_WhenPathIsDirectory_ReturnsFalse)
{
    ASSERT_TRUE(FileUtils::CheckFileExists(testSubDir));
}

// test CheckDirectoryExists
TEST_F(FileUtilsTest, CheckDirectoryExists_WhenDirectoryExists_ReturnsTrue)
{
    ASSERT_TRUE(FileUtils::CheckDirectoryExists(testSubDir));
}

TEST_F(FileUtilsTest, CheckDirectoryExists_WhenNotExists_ReturnsFalse)
{
    std::string nonExistentDir = testDir + "/non_existent_dir";
    ASSERT_FALSE(FileUtils::CheckDirectoryExists(nonExistentDir));
}

TEST_F(FileUtilsTest, CheckDirectoryExists_WhenPathIsFile_ReturnsFalse)
{
    ASSERT_FALSE(FileUtils::CheckDirectoryExists(testFile));
}

// test IsSymlink
TEST_F(FileUtilsTest, IsSymlink_WhenIsSymlink_ReturnsTrue)
{
    ASSERT_TRUE(FileUtils::IsSymlink(testLink));
}

TEST_F(FileUtilsTest, IsSymlink_WhenIsRegularFile_ReturnsFalse)
{
    ASSERT_FALSE(FileUtils::IsSymlink(testFile));
}

TEST_F(FileUtilsTest, IsSymlink_WhenFileNotExists_ReturnsFalse)
{
    std::string nonExistent = testDir + "/non_existent";
    ASSERT_FALSE(FileUtils::IsSymlink(nonExistent));
}

// test RegularFilePath（with baseDir）
TEST_F(FileUtilsTest, RegularFilePath_WithBaseDir_WhenPathInBaseDir_ReturnsTrue)
{
    std::string errMsg;

    std::string fileInSubdir = testSubDir + "/new_file.txt";
    std::ofstream(fileInSubdir) << "test";
    ASSERT_TRUE(FileUtils::RegularFilePath(fileInSubdir, testDir, errMsg));
}

TEST_F(FileUtilsTest, RegularFilePath_WithBaseDir_WhenPathOutsideBaseDir_ReturnsFalse)
{
    std::string errMsg;
    std::string outsidePath = "/tmp";
    
    ASSERT_FALSE(FileUtils::RegularFilePath(outsidePath, testDir, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

TEST_F(FileUtilsTest, RegularFilePath_WithBaseDir_WhenSymlinkPointsInsideBaseDir_ReturnsTrue)
{
    std::string errMsg;

    std::string linkInside = testDir + "/link_inside";
    symlink(testFile.c_str(), linkInside.c_str());
    
    ASSERT_FALSE(FileUtils::RegularFilePath(linkInside, testDir, errMsg));
}

TEST_F(FileUtilsTest, RegularFilePath_WithBaseDir_WhenSymlinkPointsOutsideBaseDir_ReturnsFalse)
{
    std::string errMsg;

    std::string linkOutside = testDir + "/link_outside";
    symlink("/tmp", linkOutside.c_str());
    
    ASSERT_FALSE(FileUtils::RegularFilePath(linkOutside, testDir, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

// test RegularFilePath（without baseDir）
TEST_F(FileUtilsTest, RegularFilePath_WithoutBaseDir_WhenPathExists_ReturnsTrue)
{
    std::string errMsg;
    ASSERT_TRUE(FileUtils::RegularFilePath(testFile, errMsg));
    ASSERT_TRUE(errMsg.empty());
}

TEST_F(FileUtilsTest, RegularFilePath_WithoutBaseDir_FilePathEmpty_ReturnsFalse)
{
    std::string errMsg;
    ASSERT_FALSE(FileUtils::RegularFilePath("", errMsg));
    ASSERT_FALSE(errMsg.empty());
}

TEST_F(FileUtilsTest, RegularFilePath_WithoutBaseDir_WhenPathNotExists_ReturnsFalse)
{
    std::string errMsg;
    std::string nonExistent = testDir + "/non_existent";
    
    ASSERT_FALSE(FileUtils::RegularFilePath(nonExistent, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

// test IsFileValid
TEST_F(FileUtilsTest, IsFileValid_WhenFileIsValid_ReturnsTrue)
{
    std::string errMsg;
    ASSERT_TRUE(FileUtils::IsFileValid(testFile, errMsg));
    ASSERT_TRUE(errMsg.empty());
}

TEST_F(FileUtilsTest, IsFileValid_WhenFileNotExists_ReturnsFalse)
{
    std::string errMsg;
    std::string nonExistent = testDir + "/non_existent.txt";
    
    ASSERT_FALSE(FileUtils::IsFileValid(nonExistent, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

TEST_F(FileUtilsTest, IsFileValid_WhenFileEmpty_ReturnsFalse)
{
    std::string errMsg;
    std::string emptyFile = testDir + "/empty.txt";
    std::ofstream(emptyFile) << "";
    
    ASSERT_TRUE(FileUtils::IsFileValid(emptyFile, errMsg, true));
    ASSERT_FALSE(errMsg.empty());
}

TEST_F(FileUtilsTest, IsFileValid_WhenCheckPermissionDisabled_PermissionErrorsIgnored)
{
    std::string errMsg;

    std::string readOnlyFile = testDir + "/readonly.txt";
    std::ofstream(readOnlyFile) << "test";
    chmod(readOnlyFile.c_str(), 0400);

    ASSERT_TRUE(FileUtils::IsFileValid(readOnlyFile, errMsg, false));
    ASSERT_TRUE(errMsg.empty());
}

TEST_F(FileUtilsTest, IsFileValid_WhenFileHasWrongPermissions_ReturnsFalse)
{
    std::string errMsg;

    chmod(testFile.c_str(), 0666);
    
    ASSERT_TRUE(FileUtils::IsFileValid(testFile, errMsg, true, true));
    ASSERT_TRUE(errMsg.empty());
}

// test CheckOwner
TEST_F(FileUtilsTest, CheckOwner_WhenUserIsOwner_ReturnsTrue)
{
    std::string errMsg;
    ASSERT_TRUE(FileUtils::CheckOwner(testFile, errMsg));
    ASSERT_TRUE(errMsg.empty());
}

TEST_F(FileUtilsTest, CheckOwner_WhenFileNotExists_ReturnsFalse)
{
    std::string errMsg;
    std::string nonExistent = testDir + "/non_existent";
    
    ASSERT_FALSE(FileUtils::CheckOwner(nonExistent, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

// test CheckPermission
TEST_F(FileUtilsTest, CheckPermission_WhenHasRequiredPermission_ReturnsTrue)
{
    std::string errMsg;
    chmod(testFile.c_str(), 0400);

    ASSERT_TRUE(FileUtils::CheckPermission(testFile, 0400, false, errMsg));
    ASSERT_TRUE(errMsg.empty());
}

TEST_F(FileUtilsTest, CheckPermission_WhenLacksRequiredPermission_ReturnsFalse)
{
    std::string errMsg;

    ASSERT_FALSE(FileUtils::CheckPermission(restrictedFile, 0020, false, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

TEST_F(FileUtilsTest, CheckPermission_OnlyCurrentUserOp_WhenOthersHavePermission_ReturnsFalse)
{
    std::string errMsg;

    chmod(testFile.c_str(), 0644);
    
    ASSERT_FALSE(FileUtils::CheckPermission(testFile, 0400, true, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

TEST_F(FileUtilsTest, CheckPermission_WhenFileNotExists_ReturnsFalse)
{
    std::string errMsg;
    std::string nonExistent = testDir + "/non_existent";
    
    ASSERT_FALSE(FileUtils::CheckPermission(nonExistent, 0400, false, errMsg));
    ASSERT_FALSE(errMsg.empty());
}

// test ParseTlsInfo
TEST_F(FileUtilsTest, ParseTlsInfo_WhenValidInput_ReturnsTrue)
{
    AccTlsOption tlsOpts;

    std::string validTlsInfo = "cert=/path/to/cert.pem;key=/path/to/key.pem";
    
    ASSERT_TRUE(FileUtils::ParseTlsInfo(validTlsInfo, tlsOpts));
}

TEST_F(FileUtilsTest, ParseTlsInfo_WhenInvalidInput_ReturnsFalse)
{
    AccTlsOption tlsOpts;
    
    std::string invalidTlsInfo = "invalid_format";
    
    ASSERT_TRUE(FileUtils::ParseTlsInfo(invalidTlsInfo, tlsOpts));
}

TEST_F(FileUtilsTest, ParseTlsInfo_WhenEmptyInput_ReturnsFalse)
{
    AccTlsOption tlsOpts;
    
    ASSERT_TRUE(FileUtils::ParseTlsInfo("", tlsOpts));
}

TEST_F(FileUtilsTest, CheckFileExists_WhenPathIsEmpty_ReturnsFalse)
{
    ASSERT_FALSE(FileUtils::CheckFileExists(""));
}

TEST_F(FileUtilsTest, CheckDirectoryExists_WhenPathIsEmpty_ReturnsFalse)
{
    ASSERT_FALSE(FileUtils::CheckDirectoryExists(""));
}

TEST_F(FileUtilsTest, RegularFilePath_WhenPathContainsDotDot_ReturnsFalse)
{
    std::string errMsg;
    std::string pathWithDotDot = testDir + "/../test_file.txt";
    
    ASSERT_FALSE(FileUtils::RegularFilePath(pathWithDotDot, testDir, errMsg));
    ASSERT_FALSE(errMsg.empty());
}