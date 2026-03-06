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


def generate_prob_graph(input_data: dict, name_to_func: dict):
    """
    node: (file, function)
    """
    G = nx.DiGraph()
    edges = []
    for file, file_info in input_data.items():
        include_file_list = file_info.get('includes', [])
        include_file_list_str = ''.join(include_file_list)
        for function_name, function_info in file_info.get('functions', {}).items():
            G.add_node((file, function_name))
            for callee in function_info.get('calls', []):
                callee_info = name_to_func.get(callee, [])
                for i in callee_info:
                    G.add_node(i)
                    callee_file = i[0].split('/')[-1].strip('.c').strip('.cc').strip('.py')
                    if re.search(callee_file, include_file_list_str) or 'py' in i[0]:
                        edges.append(((file, function_name), i))
    G.add_edges_from(edges)
    return G


def generate_errorcode_graph(json_data: dict, name_to_func_data: dict, all_config_data: dict, chain_dict: dict):
    G = nx.DiGraph()
    all_error_code_dict = all_config_data['knowledge-repository']
    edges = []
    for error_code, error_code_info in all_error_code_dict.items():
        if error_code != 'AISW_CANN_ERRCODE_Custom':
            G.add_node(error_code)
        rule_info = error_code_info.get('rule', [])
        for rule_dict in rule_info:
            if len(rule_dict.keys()) > 1:
                tar_code = rule_dict.get('dst_code', '')
                expression = rule_dict.get('expression', '')
                src_code = expression.strip('src.event_code == ').strip("'")
                edges.append((src_code, tar_code))
            else:
                tar_code = rule_dict.get('dst_code', '')
                if tar_code:
                    edges.append((error_code, tar_code))
    for src_code, src_code_info in chain_dict.items():
        for i in src_code_info:
            edges.append((src_code, i))
    G.add_edges_from(edges)
    return G


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


def generate_name_to_path(dependency_data: dict) -> dict:
    name_to_path = defaultdict(list)
    for fpath in dependency_data.keys():
        file_name = fpath.split('/')[-1]
        name_to_path[file_name].append(fpath)
    return name_to_path


def generate_name_to_func_data(dependency_data: dict) -> dict:
    name_to_funcs = defaultdict(list)
    for fpath, finfo in dependency_data.items():
        for fname in finfo["functions"].keys():
            key = (fpath, fname)
            name_to_funcs[fname].append(key)
    return name_to_funcs


def process_all_errorcode(all_config_data: dict, json_data: dict, output_path: str) -> dict:
    all_errorcode_data = all_config_data['knowledge-repository']
    errorcode_match_dict = {}
    count = 0
    for errorcode in all_errorcode_data.keys():
        match_result = parser_via_errorcode(errorcode, json_data)
        errorcode_match_dict[errorcode] = match_result
        if match_result:
            count += 1
    print('the average courage is ', count/len(all_errorcode_data.keys()))
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(errorcode_match_dict, f, indent=4)
        f.close()
    return errorcode_match_dict


def extract_all_sub_graph(nx_graph, errorcode_match_dict: dict, output_file_path: str):
    sub_graph = nx.DiGraph()
    errorcode_list = list(errorcode_match_dict.keys())
    chain_dict = defaultdict(list)
    for u in errorcode_list:
        for v in errorcode_list:
            if u != v:
                for i in errorcode_match_dict[u]:
                    for j in errorcode_match_dict[v]:
                        try:
                            path = nx.shortest_path(nx_graph, i, j, weight=None)
                            nx.add_path(sub_graph, path)
                            print('find chain from ', v, ' to ', u, ' : ', list(path))
                            check_accuracy(v, u)
                            chain_dict[v].append(u)
                        except Exception as e:
                            pass
    with open(output_file_path, 'w', encoding='utf-8') as f:
        json.dump(chain_dict, f, indent=4)
        f.close()
    return sub_graph, chain_dict


def chain_completion(all_config_data: dict, json_data: dict, output_path: str, output_chain_file_path: str, nx_graph, re_calculate=True):
    if re_calculate:
        match_result = process_all_errorcode(all_config_data, json_data, output_path)
    else:
        match_result = read_json_file([output_path])
    sub_graph, chain_dict = extract_all_sub_graph(nx_graph, match_result, output_chain_file_path)
    return sub_graph, chain_dict, match_result


def check_accuracy(caller: str, callee: str) -> None:
    config_file = read_json_file([ASCEND_KG_CONFIG_PATH])
    all_code = config_file['knowledge-repository']
    callee_info = all_code.get(callee, {})
    dst_code_list = callee_info.get('rule', [])
    for i in dst_code_list:
        if i.get('dst_code', '') == caller:
            print('confirmed ', callee, ' to ', caller)
    return


def main():
    json_file_path_list = []
    parsed_file_directory = ""
    json_data = read_json_file(json_file_path_list)
    name_to_func_data = generate_name_to_func_data(json_data)
    name_to_path = generate_name_to_path(json_data)
    nx_graph = generate_prob_graph(json_data, name_to_func_data)
    error_data_list = read_log_json_file(parsed_file_directory)
    all_config_data = read_json_file([ASCEND_KG_CONFIG_PATH])
    all_sub_graph, chain_dict, match_result = chain_completion(all_config_data, json_data, '../output/output.json',
                                     '../output/output_chain.json', nx_graph, re_calculate=True)
    errorcode_graph = generate_errorcode_graph(json_data, name_to_func_data, all_config_data, chain_dict)
    nx.write_graphml(nx_graph, "../output/nx_graph.graphml")
    nx.write_graphml(errorcode_graph, "../output/errorcode_graph.graphml")


if __name__ == "__main__":
    main()

