#!/usr/bin/python

# This script is used to generate json files with testing data for both the grpc server and the client

import json

server_file_path = '../testdata/test_data_server.json'
client_file_path = '../testdata/test_data_client.json'

svc_list = ['A', 'B', 'C', 'D']
svc_info_server = []

for l in svc_list:
    svc_info_server.append(dict(svc_name='svc'+l,
                                svc_desc='This is service '+l,
                                svc_pos=0))

svc_chain_list = [[0, 1, 2], [1, 3, 2], [0, 2, 3, 1]]
svc_chain_client = []

i = 0
for c in svc_chain_list:
    sc = dict()
    sc['chain_id'] = i + 1
    sc['chain_len'] = len(c)
    sc['chain'] = []
    i = i + 1
    j = 0
    for s in c:
        sc['chain'].append(dict(svc_name='svc'+svc_list[s],
                                svc_pos=j+1))
        j = j + 1
    svc_chain_client.append(sc)

with open(server_file_path, 'w') as sf:
    json.dump(svc_info_server, sf, indent=4)

with open(client_file_path, 'w') as cf:
    json.dump(svc_chain_client, cf, indent=4)
