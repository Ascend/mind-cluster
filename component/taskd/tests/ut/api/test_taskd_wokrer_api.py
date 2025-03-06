import unittest
from unittest.mock import patch

from taskd.api.taskd_worker_api import init_taskd_worker


class WorkerTestCase(unittest.TestCase):
    def test_init_taskd_worker_success(self, mock_worker):
        rank_id = 'not_an_int'
        upper_limit = 5000
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertEqual(result, False)
        rank_id = 1
        upper_limit = 'not_an_int'
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertEqual(result, False)
        rank_id = -1
        upper_limit = 500
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertEqual(result, False)
        rank_id = 1
        upper_limit = 400
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertEqual(result, False)


if __name__ == '__main__':
    unittest.main()
