#!/bin/bash

function get_eth_name_by_hca_name() {
	hca=$1
	eth_name=$(ls /sys/class/infiniband/${hca}/device/net/ 2>/dev/null || echo "")
	echo $eth_name
}

function get_hic_name() {
	hca=$1
	eth_name=$(get_eth_name_by_hca_name $hca)
	if [[ -z "$eth_name" ]]; then
		echo ""
		return
	fi
	hinicadm5 info 2>/dev/null | grep -B4 $eth_name | grep "|----.*CAL_" | grep -oP "(?<=\-\-\-\-)(.*)(?=\()" || echo ""
}

function get_host_topo_output() {
	hic_name=$1
	if [[ -z "$hic_name" ]]; then
		echo ""
		return
	fi
	hinicadm5 host_topo -i $hic_name -a 2>/dev/null
}

function get_port_width_info() {
	output=$1
	port_num=$2

	link_line=$(echo "$output" | grep -A5 "core.*port $port_num" | grep "link: width")
	if [[ -z "$link_line" ]]; then
		echo "0 0"
		return
	fi

	cap=$(echo "$link_line" | grep -oP "width:\(cap 0x\K[0-9a-fA-F]+")
	cur=$(echo "$link_line" | grep -oP "width:\(cap 0x[0-9a-fA-F]+,cur 0x\K[0-9a-fA-F]+")

	cap_dec=$((16#$cap))
	cur_dec=$((16#$cur))
	echo "$cap_dec $cur_dec"
}

function check_ub_port() {
	hca=$1

	hic_name=$(get_hic_name $hca)
	if [[ -z "$hic_name" ]]; then
		echo "false:cannot get hic name for hca $hca"
		return
	fi

	output=$(get_host_topo_output $hic_name)
	if [[ -z "$output" ]]; then
		echo "false:failed to get host_topo info for $hic_name"
		return
	fi

	result="false"
	port_down_list=""
	down_count=0

	ports=$(echo "$output" | grep -E "core.*port" | awk '{print $4}')
	total_ports=$(echo "$ports" | wc -w)
	for port in $ports; do
		port_hex=$(echo $port | sed 's/port0x//')
		port_num=$(echo $port | sed 's/port//')

		read _ cur <<< $(get_port_width_info "$output" "$port_num")
		if [[ $cur -eq 0 ]]; then
			down_count=$((down_count + 1))
			port_down_list="$port_down_list port$port_hex"
		fi
	done

	if [[ $down_count -eq $total_ports && $down_count -gt 0 ]]; then
		result="true"
	fi

	if [[ "$result" == "true" ]]; then
		echo "true:ub port down on$port_down_list, count: $down_count for hca $hca"
	else
		echo "false:all ub ports normal for hca $hca"
	fi
}

function check_ub_lane() {
	hca=$1

	hic_name=$(get_hic_name $hca)
	if [[ -z "$hic_name" ]]; then
		echo "false:cannot get hic name for hca $hca"
		return
	fi

	output=$(get_host_topo_output $hic_name)
	if [[ -z "$output" ]]; then
		echo "false:failed to get host_topo info for $hic_name"
		return
	fi

	result="false"
	lane_down_ports=""

	ports=$(echo "$output" | grep -E "core.*port" | awk '{print $4}')
	for port in $ports; do
		port_hex=$(echo $port | sed 's/port0x//')
		port_num=$(echo $port | sed 's/port//')

		read cap cur <<< $(get_port_width_info "$output" "$port_num")
		if [[ $cur -gt 0 && $cur -lt $cap ]]; then
			result="true"
			lane_down_ports="$lane_down_ports port$port_hex(cap=$cap,cur=$cur)"
		fi
	done

	if [[ "$result" == "true" ]]; then
		echo "true:lane down detected on$lane_down_ports for hca $hca"
	else
		echo "false:all lanes normal for hca $hca"
	fi
}

function check_dpu_card_drop() {
	hca=$1

	if ! command -v hinicadm5 &>/dev/null; then
		echo "false:hinicadm5 not found, cannot check card drop for hca $hca"
		return
	fi
	card_num=$(hinicadm5 info 2>/dev/null | grep -oP "Card num[^0-9]*\K[0-9]+")
	if [[ -z "$card_num" ]]; then
		echo "false:failed to get card num"
		return
	fi
	if [[ "$card_num" -eq 0 ]]; then
		echo "false:get card num = 0, skip"
		return
	fi

	for((i=0;i<$card_num;i++))
	do
		cfg_template_index=$(hinicadm5 cfg_template -i hinic$i 2>/dev/null | grep -oP "Current[^0-9]*\K[0-9]+")
		# 配置模板为如下格式：
		# "Template[ 1]: ETH_2X400GR4_4DAVID_2HOST"
		cfg_template=$(hinicadm5 cfg_template -i hinic$i 2>/dev/null | grep -oP "Template\[ ${cfg_template_index}\]: \K\S+")
		if [[ ! -z "$cfg_template" ]]; then
			break
		fi
	done
	if [[ -z "$cfg_template" ]]; then
		echo "false:failed to get cfg template"
		return
	fi
	# 配置模板说明
	# ETH_2X400GR4_UBC_4DAVID_2HOST   物理口为2个400G，每个DPU对应4个NPU
	# ETH_2X400GR4_UBC_2DAVID_1HOST   物理口为2个400G，每个DPU对应2个NPU
	# ETH_4X200GR2_UBC_4DAVID_2HOST   物理口为4个200G，每个DPU对应4个NPU
	# ETH_4X200GR2_UBC_2DAVID_1HOST   物理口为4个200G，每个DPU对应2个NPU
	# 服务器两类配置模板,主要区别在物理口速率，DPU数量不变
	# ETH_2X400GR4_UBC_2HOST_UB_EXP
	# ETH_4X200GR4_UBC_2HOST_UB_EXP
	case $cfg_template in
		ETH_2X400GR4_UBC_4DAVID_2HOST|ETH_4X200GR2_UBC_4DAVID_2HOST)
			expected_phy_num=2
			;;
		ETH_2X400GR4_UBC_2DAVID_1HOST|ETH_4X200GR2_UBC_2DAVID_1HOST)
			expected_phy_num=4
			;;
		ETH_2X400GR4_UBC_2HOST_UB_EXP|ETH_4X200GR4_UBC_2HOST_UB_EXP) # server
			expected_phy_num=4
			;;
		*) # 其他不支持机型设置为0，不检查
			expected_phy_num=0
			;;
	esac
	if [[ "$expected_phy_num" -eq 0 ]]; then
		echo "false:no card drop detected, expected_phy_num=0"
		return
	fi
	if [[ "$card_num" -ge "$expected_phy_num" ]]; then
		echo "false:no card drop detected, card num=$card_num"
	else
		echo "true:card drop detected, expected=$expected_phy_num, actual=$card_num"
	fi
}

"$@"
