# Python api for mindio_acp

use mindio speed up torch.save() torch.load()

you need to work it with memfs sevrer.

## Pytorch How to use:

1. pip install mindio_acp*.whl
2. import mindio_acp
3. mindio_acp.save(obj, PATH) same as torch.save(obj, PATH)
4. mindio_acp.load(PATH) same as torch.load(PATH)

## Pytorch example:

```python
import time
import io
import os

import torch
import mindio_acp

# torch org ckpt path
file_path = '/home/ckpt_xxx/xxx.pt'

t_start_load = time.time()
ckpt = torch.load(file_path)
t_end_load = time.time()
print('org torch.load load time:', t_end_load - t_start_load)


mem_name = "/mnt/xxx/xxx.mpt"

t_start_torch_save = time.time()
mindio_acp.save(ckpt, mem_name)
t_end_torch_save = time.time()
print('new save time by mindio:', t_end_torch_save - t_start_torch_save)

t_start_torch_load = time.time()
ckpt2 = mindio_acp.load(mem_name)
t_end_torch_load = time.time()
print('new load time by mindio:',t_end_torch_load - t_start_torch_load)

```

## MindSpore How to use:

1. pip install mindio_acp*.whl
2. import mindio_acp
3. with mindio_acp.create_file(PATH) as fd:
        fd.write(ckpt)
4. with mindio_acp.open_file(PATH) as fd:
        fd.read()

## MindSpore example:

```python
import time
import io
import os

import torch
import mindio_acp

# mindspore org ckpt path
file_path = '/home/ckpt_xxx/xxx.pt'

t_start_load = time.time()
ckpt = torch.load(file_path)
t_end_load = time.time()
print('org torch.load load time:', t_end_load - t_start_load)


mem_name = "/mnt/xxx/xxx.mpt"

t_start_torch_save = time.time()
with mindio_acp.create_file(mem_name) as fd:
    fd.write(ckpt)
t_end_torch_save = time.time()
print('new save time by mindio:', t_end_torch_save - t_start_torch_save)

t_start_torch_load = time.time()
with mindio_acp.open_file(mem_name) as fd:
    ckpt2 = fd.read()
t_end_torch_load = time.time()
print('new load time by mindio:',t_end_torch_load - t_start_torch_load)

```