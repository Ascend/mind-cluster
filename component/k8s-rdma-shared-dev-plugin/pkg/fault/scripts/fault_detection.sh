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
	hinicadmdfx5 info 2>/dev/null | grep -B4 $eth_name | grep "|----.*CAL_" | grep -oP "(?<=\-\-\-\-)(.*)(?=\()" || echo ""
}

function get_pcie_topo_output() {
	hic_name=$1
	if [[ -z "$hic_name" ]]; then
		echo ""
		return
	fi
	hinicadmdfx5 pcie_topo -i $hic_name -a 2>/dev/null
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

	output=$(get_pcie_topo_output $hic_name)
	if [[ -z "$output" ]]; then
		echo "false:failed to get pcie_topo info for $hic_name"
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

	output=$(get_pcie_topo_output $hic_name)
	if [[ -z "$output" ]]; then
		echo "false:failed to get pcie_topo info for $hic_name"
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

	if ! command -v hinicadmdfx5 &>/dev/null; then
		echo "false:hinicadmdfx5 not found, cannot check card drop for hca $hca"
		return
	fi

	eth_name=$(get_eth_name_by_hca_name $hca)
	if [[ -z "$eth_name" ]]; then
		echo "false:cannot get eth name for hca $hca"
		return
	fi

	if ! command -v hinicadmdfx5 &>/dev/null; then
		echo "false:hinicadmdfx5 not found, cannot check card drop for hca $hca"
		return
	fi

	card_type_path="/sys/class/net/$eth_name/device/card_type"
	if [[ ! -f "$card_type_path" ]]; then
		echo "false:card_type file not found at $card_type_path"
		return
	fi

	card_type=$(cat "$card_type_path" 2>/dev/null)
	phy_num=-1
	case "$card_type" in
		A5Server|A5Pod400G2david)
			phy_num=4
			;;
		A5Pod200G2david|A5Pod200G4david)
			phy_num=2
			;;
	esac

	if [[ $phy_num -lt 0 ]]; then
		echo "false:unknown card_type '$card_type'"
		return
	fi

	card_num=$(hinicadmdfx5 info 2>/dev/null | grep -oP "Card num[^0-9]*\K[0-9]+")
	if [[ -z "$card_num" ]]; then
		echo "false:failed to get card num"
		return
	fi

	if [[ "$card_num" -eq "$phy_num" ]]; then
		echo "false:no card drop detected, card num=$card_num"
	else
		echo "true:card drop detected, expected=$phy_num, actual=$card_num"
	fi
}

"$@"
