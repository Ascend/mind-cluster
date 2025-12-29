import copy
import json
import networkx as nx
import re
from collections import defaultdict
import os
import pickle
import numpy as np
from scipy.stats import fisher_exact
from datetime import datetime, timedelta

ASCEND_KG_CONFIG_PATH = '../data/ascend-kg-config.json'

def read_json_file(file_path_list: list) -> dict:
    merged_dict = {}
    for i in file_path_list:
        with open(i, 'r', encoding='utf-8') as f:
            data = json.load(f)
            merged_dict.update(data)
            f.close()
    return merged_dict


def read_log_json_file(log_file_path: str) -> list:
    log_list = []
    for filename in os.listdir(log_file_path):
        if filename.endswith(".txt") or filename.endswith(".json"):
            filepath = os.path.join(log_file_path, filename)
            with open(filepath, "r", encoding="utf-8") as f:
                data = json.load(f)
                data['file_name'] = filename
                log_list.append(data)
                f.close()
    return log_list


def sort_root_cause(root_cause_list: list, prob_graph, errorcode_graph, match_result: dict) -> list:
    """
    若故障根因大于3，则启动故障根因重排序；
    """
    return sort_via_module_code(errorcode_graph, root_cause_list)


def sort_via_module_code(errorcode_graph, rc_list: list) -> list:
    node_list = []
    for rc in rc_list:
        error_code = rc.get('code', '')
        component = rc.get('component', '')
        module = rc.get('module', '')
        node_list.append((error_code, component, module))
    # Fisher精确检验构造因果边
    sub_graph = created_related_graph(errorcode_graph, node_list, rc_list)
    print('---------------------', list(sub_graph.edges()), '--------------------------')
    personalization = generate_personalization_dict(sub_graph)
    sorted_root_cause = nx.pagerank(sub_graph, alpha=0.85, personalization=personalization)
    print("PageRank scores:")
    for node, score in sorted_root_cause.items():
        print(f"{node}: {score:.4f}")
    sorted_root_cause = dict(sorted(sorted_root_cause.items(), key=lambda x: x[1], reverse=True))
    order_index = {rank_result: idx for idx, rank_result in enumerate(sorted_root_cause)}
    sorted_rc_list = sorted(rc_list, key=lambda x: order_index.get(x['code'], float('inf')))
    return sorted_rc_list


def generate_personalization_dict(nx_graph) -> dict:
    personalization = {n: 20 for n in nx_graph.nodes()}
    for k in personalization.keys():
        if k.startswith('AISW_TRACEBACK'):
            personalization[k] -= 18
        if k.startswith('AISW_CANN_ERRCODE_Custom'):
            personalization[k] -= 5
        if k.startswith('AISW_CANN_AMCT'):
            personalization[k] -= 19
        if k.startswith('0x'):
            personalization[k] += 10
        if k.startswith('Comp'):
            personalization[k] += 0
    total = sum(personalization.values())
    personalization_dict = {n: v / total for n, v in personalization.items()}
    return personalization_dict


def created_related_graph(ori_errorcode_graph, error_node_list: list, rc_list: list):
    """
    创建新的带权重子图
    """
    G = nx.DiGraph()
    single_code_list = [c[0] for c in error_node_list]
    for i in error_node_list:
        G.add_node(i[0])
        if i[0] in ori_errorcode_graph.nodes:
            in_nodes = list(ori_errorcode_graph.predecessors(i[0]))
            for in_node in in_nodes:
                G.add_edge(i[0], in_node)
            out_nodes = list(ori_errorcode_graph.successors(i[0]))
            for out_node in out_nodes:
                G.add_edge(out_node, i[0])
    new_edge_list = find_new_edges_from_log(rc_list)
    G.add_edges_from(new_edge_list)
    return G


def find_new_edges_from_log(root_cause_list: list) -> list:
    edge_list = []
    for i in range(len(root_cause_list)):
        for j in range(len(root_cause_list)):
            if i != j:
                if root_cause_list[i]['code'].startswith('AISW_TRACEBACK') or root_cause_list[j]['code'].startswith('AISW_TRACEBACK'):
                    continue
                try:
                    src_log = process_log(root_cause_list[i])
                    tar_log = process_log(root_cause_list[j])
                    # Execution Anomaly Detection in Distributed Systems through Unstructured Log Analysis
                    table = build_contingency(src_log, tar_log, delta_ms=10000)
                    odds_ratio, p_value = fisher_exact(table)
                    if p_value < 0.05 and odds_ratio > 1:
                        print('create fisher edge from ', root_cause_list[i]['code'], ' to ', root_cause_list[j]['code'])
                        edge_list.append((root_cause_list[j]['code'], root_cause_list[i]['code']))
                except:
                    continue
    return edge_list


def process_log(errorcode_info: dict) -> list:
    log_list = []
    event_attr = errorcode_info.get('event_attr', {})
    for error_info in event_attr.values():
        for i in error_info:
            log_list.append((i.get('key_info', ''), datetime.fromisoformat(i.get('occur_time', ''))))
    return log_list


def build_contingency(log_a, log_b, delta_ms=1000):
    times_a = [t for _, t in log_a]
    times_b = [t for _, t in log_b]
    delta = timedelta(milliseconds=delta_ms)
    n11 = n10 = n01 = n00 = 0
    for t_a in times_a:
        has_b = any(t_a <= t_b <= t_a + delta for t_b in times_b)
        if has_b:
            n11 += 1
        else:
            n10 += 1
    t_min, t_max = min(times_a + times_b), max(times_a + times_b)
    total_seconds = (t_max - t_min).total_seconds()
    random_times = [t_min + timedelta(seconds=np.random.uniform(0, total_seconds))
                    for _ in range(len(times_a))]
    for t_r in random_times:
        has_b = any(t_r <= t_b <= t_r + delta for t_b in times_b)
        if has_b:
            n01 += 1
        else:
            n00 += 1
    return [[n11, n10], [n01, n00]]


def parse_rc_list(root_cause_list: list, match_result: dict) -> list:
    node_list = []
    for rc in root_cause_list:
        first_device_data = rc.get('event_attr', {})
        error_code = rc.get('code', '')
        if first_device_data:
            first_device, first_device_info = next(iter(first_device_data.items()))
            location_list = match_result.get(error_code, [])

            for i in location_list:
                node_list.append((first_device_info[0].get('event_code', ''), i[0], i[1]))
    return node_list


def parser_via_errorcode(error_code: str, json_data: dict) -> list:
    """
    函数func名
    """
    location_list = []
    config_file = read_json_file([ASCEND_KG_CONFIG_PATH])
    config_file = config_file['knowledge-repository']
    error_code_info = config_file.get(error_code, {})
    error_code_regex = error_code_info.get('regex', {})
    error_code_regex_in = error_code_regex.get('in', [])
    if not error_code_regex_in:
        return []
    if isinstance(error_code_regex_in[0], list):
        for i in error_code_regex_in:
            match_result = match_function(i, json_data)
            if match_result[0]:
                print('match result is', error_code, match_result)
                location_list.append(match_result)
    else:
        match_result = match_function(error_code_regex_in, json_data)
        if match_result[0]:
            print('match result is', error_code, match_result)
            location_list.append(match_function(error_code_regex_in, json_data))
    return location_list


def match_function(regex_list: list, json_data: dict) -> tuple:
    for f_path, v in json_data.items():
        functions_info = v.get('functions', {})
        for f_name, f_info in functions_info.items():
            f_code = f_info.get('code', '')
            count = 0
            for regex in regex_list:
                if regex not in f_code:
                    break
                else:
                    count += 1
            if len(regex_list) == count:
                return (f_path, f_name)
    return ('', '')


def read_root_cause(parsed_error_dict: dict) -> list:
    knowledge_graph = parsed_error_dict.get('Knowledge_Graph', {})
    root_cause_list = knowledge_graph.get('fault', [])
    new_root_cause_list = process_duplicated_code(root_cause_list)
    return new_root_cause_list


def process_duplicated_code(rc_list: list) -> list:
    """
    处理诊断报告中出现重复的code的情况，将重复的code中的event_attr合并
    """
    new_dict_list = []
    code_index_dict = {}
    count = 0
    for i in rc_list:
        if i['code'] in code_index_dict.keys():
            code_index = code_index_dict[i['code']]
            event = i['event_attr']
            for k, v in event.items():
                new_dict_list[code_index]['event_attr'][k] = v
        else:
            new_dict_list.append(i)
            code_index_dict[i['code']] = count
            count += 1
    return new_dict_list


def main():
    # 处理故障日志
    parsed_file_directory = r''
    error_data_list = read_log_json_file(parsed_file_directory)
    nx_graph = nx.read_graphml("../output/nx_graph.graphml")
    errorcode_graph = nx.read_graphml("../output/errorcode_graph.graphml")
    with open('../output/output.json', 'r', encoding='utf-8') as f:
        match_result = json.load(f)

    new_error_data = dict()
    all_file_count = 0
    file_count = 0
    correct_count = 0
    for error_data in error_data_list:
        new_error_data = copy.deepcopy(error_data)
        root_cause_list = read_root_cause(error_data)
        if len(root_cause_list) > 3 and "TRACEBACK" not in error_data['file_name']:
            result = sort_root_cause(root_cause_list, nx_graph, errorcode_graph, match_result)
            print('************  file ', error_data['file_name'], '  ***********')
            print('ori root cause list is ', [i['code'] for i in root_cause_list])
            print('new root cause list is ', [i['code'] for i in result])
            all_file_count += 1
            if root_cause_list != result:
                file_count += 1
                new_error_data['fault'] = result
            root_cause = error_data['file_name'].strip('.txt').strip('.json').split('-')[-1]
            if root_cause in [i['code'] for i in result][:3]:
                correct_count += 1
            else:
                print('error file name is ', error_data['file_name'])
    print('final rerank file num is ', file_count)
    print('accuracy is ', correct_count, '/', all_file_count, '=', correct_count/all_file_count)
    return new_error_data


if __name__ == "__main__":
    main()

