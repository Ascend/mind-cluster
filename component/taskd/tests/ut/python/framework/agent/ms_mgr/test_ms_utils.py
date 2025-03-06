import unittest
from unittest.mock import patch

from taskd.python.framework.agent.ms_mgr.ms_utils import check_monitor_res_valid, calculate_global_rank
from taskd.python.utils.log import run_log


class TestFunctions(unittest.TestCase):

    def test_check_monitor_res_valid_valid_input(self):
        rank_status_dict = {
            1: {'pid': 1, 'status': 0, 'global_rank': 1},
            2: {'pid': 2, 'status': 1, 'global_rank': 2}
        }
        result = check_monitor_res_valid(rank_status_dict)
        self.assertEqual(result, True)

    def test_check_monitor_res_valid_non_dict_input(self):
        rank_status_dict = [1, 2, 3]
        result = check_monitor_res_valid(rank_status_dict)
        self.assertEqual(result, False)

    def test_check_monitor_res_valid_non_dict_info(self):
        rank_status_dict = {
            'rank1': [1, 2, 3]
        }
        result = check_monitor_res_valid(rank_status_dict)
        self.assertEqual(result, False)

    def test_check_monitor_res_valid_missing_key(self):
        rank_status_dict = {
            'rank1': {'pid': 1, 'status': 0}
        }
        result = check_monitor_res_valid(rank_status_dict)
        self.assertEqual(result, False)

    def test_check_monitor_res_valid_non_int_pid(self):
        rank_status_dict = {
            'rank1': {'pid': 'not_an_int', 'status': 0, 'global_rank': 1}
        }
        result = check_monitor_res_valid(rank_status_dict)
        self.assertEqual(result, False)

    def test_check_monitor_res_valid_non_int_status(self):
        rank_status_dict = {
            'rank1': {'pid': 1, 'status': 'not_an_int', 'global_rank': 1}
        }
        result = check_monitor_res_valid(rank_status_dict)
        self.assertEqual(result, False)

    def test_check_monitor_res_valid_non_int_global_rank(self):
        rank_status_dict = {
            'rank1': {'pid': 1, 'status': 0, 'global_rank': 'not_an_int'}
        }
        result = check_monitor_res_valid(rank_status_dict)
        self.assertEqual(result, False)

    @patch('os.getenv')
    def test_calculate_global_rank_valid_input(self, mock_getenv):
        mock_getenv.side_effect = ['2', '3']
        result = calculate_global_rank()
        expected = [6, 7]
        self.assertEqual(result, expected)

    @patch('os.getenv')
    def test_calculate_global_rank_missing_env_variable(self, mock_getenv):
        mock_getenv.return_value = None
        result = calculate_global_rank()
        self.assertEqual(result, [])

    @patch('os.getenv')
    def test_calculate_global_rank_invalid_env_variable(self, mock_getenv):
        mock_getenv.side_effect = ['not_an_int', 'not_an_int']
        result = calculate_global_rank()
        self.assertEqual(result, [])

if __name__ == '__main__':
    unittest.main()