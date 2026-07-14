# Solution Introduction<a name="ZH-CN_TOPIC_0000002479386900"></a>

Volcano can implement affinity scheduling in the following two aspects: Ascend AI processor-based affinity and node-based affinity.

## Basic Concepts<a name="section988510210395"></a>

- Ascend AI processor-based affinity
    - Ascend AI processor-based affinity rule: Based on the interconnection topology and processing logic of Ascend AI processors, this rule enables the best use of chips.
    - Affinity scheduling policy: Based on the affinity rules of Ascend AI processors, it implements the scheduling logic for Volcano to select specific Ascend AI processors. Based on the affinity scheduling policy and scheduling principles, optimal resource allocation can be achieved.

- Node-based affinity
    - Switch affinity scheduling: Based on the networking configuration and parameter plane network configuration of nodes under a switch, it achieves the best use of nodes.
    - Logical SuperPoD affinity scheduling: Physical SuperPoDs in the cluster are divided into logical SuperPoDs according to the splitting policy, achieving the best use of nodes.
    - Frame affinity scheduling: Atlas 950 SuperPoD consists of multiple physical frames, each with 8 nodes. The network communication performance between nodes within a frame is better, achieving the best use of nodes.

## Ascend AI Processor-Based Affinity <a name="section18208162194419"></a>

This document details the Ascend AI processor-based affinity rules for products such as the Atlas training series, Atlas 200T A2 Box16 heterogeneous subrack, A200T A3 Box8 SuperPoD, Atlas 350 PCIe card, Atlas 850 hardware products, and Atlas 950 SuperPoD, as well as Volcano scheduling rules developed on this basis.

## Node-Based Affinity <a name="section654613453444"></a>

This document also introduces the affinity rules for products such as the Atlas training series, Atlas A2 training series, Atlas 900 A3 SuperPoD, and Atlas 950 SuperPoD, that is, the node scheduling rules for switches; it provides a detailed introduction to selecting which node under a switch to invoke in the Spine-Leaf networking mode.
