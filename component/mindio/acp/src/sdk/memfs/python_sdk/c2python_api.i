/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */

%module(docstring="c to python API", directors=1, threads=1, naturalvar=1) c2python_api

#define SWIGWORDSIZE64

%{
#include "c2python_api.h"
#include "../sdk/memfs_sdk_types.h"

using namespace ock::memfs;
using namespace ock::memfs::sdk;
%}

%include "stl.i"
%include "stdint.i"
%include "typemaps.i"
%include "std_unordered_map.i"
%include "std_map.i"

%feature("director") CallbackReceiver;

%inline %{
    typedef long int off_t;
    typedef unsigned int mode_t;
%}

%template() std::vector<std::string>;
%template() std::unordered_map<std::string, std::string>;
%template() std::map<std::string, std::string>;
%template() std::unordered_map<std::string, uint64_t>;

%typemap(out) ssize_t {
    $result = PyLong_FromLong($1);
}

%typemap(in) void* {
    // 将Python对象转换为C/C++中的const void *指针
    if (PyBytes_Check($input)) {
        // 获取bytes对象的数据
        $1 = (void *)PyBytes_AsString($input);
    } else {
        PyErr_SetString(PyExc_TypeError, "A bytes object is required");
        SWIG_fail;
    }
}

%typemap(in) std::vector<ock::memfs::Buffer> {
    // 检查输入是否为列表
    if (!PyList_Check($input)) {
        PyErr_SetString(PyExc_TypeError, "Expected a list");
        SWIG_fail;
    }
    // 创建一个空的vector
    std::vector<ock::memfs::Buffer> bufferVec;
    // 填充vector
    for (int i = 0; i < PyList_Size($input); i++) {
        PyObject *py_item = PyList_GetItem($input, i);
        if (py_item == nullptr) {
            PyErr_SetString(PyExc_TypeError, "Invalid write list param");
            SWIG_fail;
        }
        if (!PyTuple_Check(py_item) || PyTuple_Size(py_item) != 2) {
            PyErr_SetString(PyExc_TypeError, "Expected a tuple of (bytes, int)");
            SWIG_fail;
        }
        PyObject *buf = PyTuple_GetItem(py_item, 0);
        PyObject *bufSize = PyTuple_GetItem(py_item, 1);
        if (buf == nullptr || bufSize == nullptr) {
            PyErr_SetString(PyExc_TypeError, "Invalid write list element");
            SWIG_fail;
        }
        if (!PyBytes_Check(buf) && !PyLong_Check(buf)) {
            PyErr_SetString(PyExc_TypeError, "First element of tuple must be bytes or pointer address");
            SWIG_fail;
        }
        if (!PyLong_Check(bufSize)) {
            SWIG_exception(SWIG_TypeError, "Second element of tuple must be an integer");
            SWIG_fail;
        }
        if (i == 0 || i == PyList_Size($input) - 1) {
            ock::memfs::Buffer buffer{ PyBytes_AsString(buf), PyLong_AsUnsignedLongLong(bufSize) };
            bufferVec.emplace_back(buffer);
        } else {
            ock::memfs::Buffer buffer{ PyLong_AsVoidPtr(buf), PyLong_AsUnsignedLongLong(bufSize) };
            bufferVec.emplace_back(buffer);
        }
    }
    $1 = bufferVec;
}

%typemap(in) std::vector<ock::memfs::ReadBuffer> {
    // 检查输入是否为列表
    if (!PyList_Check($input)) {
        PyErr_SetString(PyExc_TypeError, "Expected a list");
        SWIG_fail;
    }
    // 创建一个空的vector
    std::vector<ock::memfs::ReadBuffer> bufferVec;
    // 填充vector
    for (int i = 0; i < PyList_Size($input); i++) {
        PyObject *py_item = PyList_GetItem($input, i);
        if (py_item == nullptr) {
            PyErr_SetString(PyExc_TypeError, "Invalid read list param");
            SWIG_fail;
        }
        if (!PyList_Check(py_item) || PyList_Size(py_item) != 3) {
            PyErr_SetString(PyExc_TypeError, "Expected a list of (bytes, int, int)");
            SWIG_fail;
        }
        PyObject *buf = PyList_GetItem(py_item, 0);
        PyObject *bufStart = PyList_GetItem(py_item, 1);
        PyObject *bufSize = PyList_GetItem(py_item, 2);
        if (buf == nullptr || bufStart == nullptr || bufSize == nullptr) {
            PyErr_SetString(PyExc_TypeError, "Invalid write list element");
            SWIG_fail;
        }
        if (!PyBytes_Check(buf) && !PyLong_Check(buf)) {
            PyErr_SetString(PyExc_TypeError, "First element of tuple must be bytes or pointer address");
            SWIG_fail;
        }
        if (!PyLong_Check(bufStart) || !PyLong_Check(bufSize)) {
            SWIG_exception(SWIG_TypeError, "Second and third element of tuple must be an integer");
            SWIG_fail;
        }
        ock::memfs::ReadBuffer buffer{ PyLong_AsVoidPtr(buf),
                                       static_cast<uint64_t>(PyLong_AsLong(bufStart)),
                                       static_cast<uint64_t>(PyLong_AsLong(bufSize)) };
        bufferVec.emplace_back(buffer);
    }
    $1 = bufferVec;
}

%typemap(in) PyObject* {
    $1 = $input;
    Py_XINCREF($1);
}

%rename("%(undercase)s") "";
namespace ock::memfs::sdk {
%rename(link) LinkFile;
%rename(preload) PreloadFile;
%rename(check_background_task) CheckBackgroundTask;

// read
%rename(open) ReadableFile::Initialize;
%rename(open) NdsReadableFile::Initialize;
%rename(open) FReadableFile::Initialize;
%rename(multi_read) Readable::ReadV;
%rename(multi_read) ReadableFile::ReadV;
%rename(multi_read) NdsReadableFile::ReadV;
%rename(multi_read) FReadableFile::ReadV;
%rename(size) Readable::Length;
%rename(size) ReadableFile::Length;
%rename(size) NdsReadableFile::Length;
%rename(size) FReadableFile::Length;

// write
%rename(create) WriteableFile::Initialize;
%rename(create) FWriteableFile::Initialize;
%rename(write_list) Writeable::WriteV;
%rename(write_list) WriteableFile::WriteV;
%rename(write_list) FWriteableFile::WriteV;
}

%include "c2python_api.h"