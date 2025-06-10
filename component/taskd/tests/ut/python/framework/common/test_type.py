#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
import unittest
from taskd.python.framework.common.type import Position, TLSConfig, NetworkConfig, \
     MessageInfo, MsgBody, AgentReportInfo


class TestDataClasses(unittest.TestCase):
    def setUp(self):
        self.position = Position(role='test_role', server_rank='0', process_rank='0')
        self.tls_config = TLSConfig(
            ca='ca_path',
            server_key='server_key_path',
            server_crt='server_crt_path',
            client_key='client_key_path',
            client_crt='client_crt_path'
        )
        self.network_config = NetworkConfig(
            pos=self.position,
            upstream_addr='upstream_addr',
            listen_addr='listen_addr',
            enable_tls=True,
            tls_conf=self.tls_config
        )
        self.message_info = MessageInfo(
            uuid='test_uuid',
            biz_type='test_biz_type',
            dst=self.position,
            body='test_body'
        )
        self.msg_body = MsgBody(
            msg_type='test_msg_type',
            code=0,
            message='test_message',
            extension={'key': 'value'}
        )
        self.agent_report_info = AgentReportInfo(
            fault_ranks=[1, 1, 1],
            restart_times=1
        )

    def test_position(self):
        self.assertEqual(self.position.role, 'test_role')
        self.assertEqual(self.position.server_rank, '0')
        self.assertEqual(self.position.process_rank, '0')

    def test_tls_config(self):
        self.assertEqual(self.tls_config.ca, 'ca_path')
        self.assertEqual(self.tls_config.server_key, 'server_key_path')
        self.assertEqual(self.tls_config.server_crt, 'server_crt_path')
        self.assertEqual(self.tls_config.client_key, 'client_key_path')
        self.assertEqual(self.tls_config.client_crt, 'client_crt_path')

    def test_network_config(self):
        self.assertEqual(self.network_config.pos, self.position)
        self.assertEqual(self.network_config.upstream_addr, 'upstream_addr')
        self.assertEqual(self.network_config.listen_addr, 'listen_addr')
        self.assertTrue(self.network_config.enable_tls)
        self.assertEqual(self.network_config.tls_conf, self.tls_config)

    def test_message_info(self):
        self.assertEqual(self.message_info.uuid, 'test_uuid')
        self.assertEqual(self.message_info.biz_type, 'test_biz_type')
        self.assertEqual(self.message_info.dst, self.position)
        self.assertEqual(self.message_info.body, 'test_body')

    def test_msg_body(self):
        self.assertEqual(self.msg_body.msg_type, 'test_msg_type')
        self.assertEqual(self.msg_body.code, 0)
        self.assertEqual(self.msg_body.message, 'test_message')
        self.assertEqual(self.msg_body.extension, {'key': 'value'})

    def test_agent_report_info(self):
        self.assertEqual(self.agent_report_info.fault_ranks, [1, 1, 1])
        self.assertEqual(self.agent_report_info.restart_times, 1)


if __name__ == '__main__':
    unittest.main()
