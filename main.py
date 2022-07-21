import random
import argparse
import pandas as pd
import math
import numpy as np
import os
from utils import generate_gln_list, cte1, cte2, cte3, cte4, cte5, cte5_repack, cte6, cte7

# 1: fishing company/vessel, 2: auction center; 3: logistic service provider; 4: processing factory;
# 5: distribution center/wholesaler; 6: retailer
pis = "123436"

# 0 : gtin remains the same as before
# n (n>=1): number of input gtins (in cte4 processing event or cte5 packing event) at merging point
merge_gtin = "000100"
# 0 : gtin remains the same as before
# n (n>=1): number of output gtins (cte4 or cte5) at splitting point
split_gtin = "000200"
# n (n>1): number of customers (cte2 at auction center or cte6 at processing factory/distribution center),
# split path without changing gtin
split_pi = "000000"


parser = argparse.ArgumentParser(description='Data synthesis')
# parser.add_argument('--pis', type=int, default=6, help='Merge paths')
parser.add_argument('--merge_gtin', type=str, default=merge_gtin, help='Control the merge')
parser.add_argument('--split_gtin', type=str, default=split_gtin, help='Control the splitting')
parser.add_argument('--split_pi', type=str, default=split_pi, help='Control the splitting (without changing gtin)')
parser.add_argument('--pis', type=str, default=pis, help='Point of interests in the path')
parser.add_argument('--sample_num', type=int, default=20, help='Number of paths')
parser.add_argument('--same_pis', type=bool, default=False,
                    help='receive several gtins from the same point of interest or not')
args = parser.parse_args()


def save_to_csv(list, path):
    df = pd.DataFrame(list)
    df.to_csv(path, index=False)


def generate(pis, gap, pi_glns, same_flag, path_data):
    """
    :param pis: point of interests in the path; the number represents the role of PI in the path.
    :param gap: starting from the gap, the data will be generated.
    :param pi_glns: gln list for each role of PI.
    :param same_flag: whether receive several gtins from the same PI or split gtin and send to the same
    PI, only valid when merging or splitting.
    :param path_data: a dictionary to store the generated data.
    :return:
    """
    for i in range(0+gap, len(pis)):
        if int(pis[i]) == 1:
            cte1_data = []
            if len(path_data[0])!= 0 and same_flag == True:
                vessel_gln = path_data[0][0]['vessel_gln']
                next_pi_gln = path_data[0][0]['next_pi_gln']
            else:
                vessel_gln = random.choice(tuple(pi_glns[0]))
                next_pi_gln = random.choice(tuple(pi_glns[int(args.pis[1])-1]))
            cte1_data += cte1(vessel_gln, 1)
            cte1_data[-1]['generator_gln'] = vessel_gln
            cte1_data[-1]['next_pi_gln'] = next_pi_gln
            path_data[0] += cte1_data
        elif int(pis[i]) == 2:
            cte2_data = []
            split_num = 1 if (int(split_pi[i]) < 2) else int(split_pi[i])
            for j in range(0, split_num):
                last_event = path_data[0][-1]
                auction_gln = last_event['next_pi_gln']
                if len(path_data[i]) != 0:
                    customer_gln = path_data[i][0]['customer_gln']
                    if same_flag:
                        next_pi_gln = path_data[i][0]['next_pi_gln']
                else:
                    next_pi = int(args.pis[i+1])
                    # auction_gln = last_event['next_pi_gln']
                    if next_pi == 3:
                        next_pi_gln = random.choice(tuple(pi_glns[2]))
                        customer_gln = random.choice(tuple(pi_glns[int(args.pis[i+2]) - 1]))
                    else:
                        customer_gln = random.choice(tuple(pi_glns[next_pi - 1]))
                        next_pi_gln = customer_gln
                weight = int(last_event['weight'] / split_num)
                cte2_data += cte2(last_event, auction_gln, customer_gln, weight)
                cte2_data[-1]['generator_gln'] = auction_gln
                cte2_data[-1]['last_pi_gln'] = last_event['vessel_gln']
                cte2_data[-1]['next_pi_gln'] = next_pi_gln
            path_data[i] += cte2_data
        elif int(pis[i]) == 3:
            cte3_data = []
            split_pi_num = 1 if (int(split_pi[i-1]) < 2) else int(split_pi[i-1])
            split_gtin_num = 1 if (int(split_gtin[i-1]) < 2) else int(split_gtin[i-1])
            split_num = split_pi_num * split_gtin_num
            for j in range(0, split_num):
                if isinstance(path_data[i - 1][-1], dict):
                    last_event = path_data[i - 1][-1] if (split_num == 1) else path_data[i - 1][j]
                else:
                    try:
                        last_event = path_data[i - 1][-1][j]
                    except IndexError:
                        print(path_data[i - 1][-1])
                carrier_gln = last_event['next_pi_gln']
                cte3_data += cte3(last_event, carrier_gln)
                cte3_data[j]['generator_gln'] = carrier_gln
                cte3_data[j]['last_pi_gln'] = last_event['generator_gln']
                cte3_data[j]['next_pi_gln'] = last_event['customer_gln']
                # print(cte3_data[j])
            path_data[i] += cte3_data

        elif int(pis[i]) == 4:
            cte4_data = []
            cte5_data = []
            cte6_data = []
            factory_gln = path_data[i - 1][-1]['customer_gln']

            input_gtins = []
            last_pi_gln = []
            for event in path_data[i - 1]:
                input_gtins.append(event['gtin'])
                last_pi_gln.append(event['generator_gln'])
            output_num = 1 if (int(split_gtin[i]) < 2) else int(split_gtin[i])
            cte4_data += cte4(factory_gln, input_gtins, output_num)
            cte4_data[-1]['generator_gln'] = factory_gln
            cte4_data[-1]['last_pi_gln'] = last_pi_gln
            cte4_data[-1]['next_pi_gln'] = factory_gln

            for event in cte4_data:
                factory_gln = event['generator_gln']
                cte5_data += cte5(event, factory_gln)
                cte5_data[-1]['generator_gln'] = factory_gln
                cte5_data[-1]['last_pi_gln'] = factory_gln
                cte5_data[-1]['next_pi_gln'] = factory_gln

            for event in cte5_data:
                carrier_gln = random.choice(tuple(pi_glns[2]))
                customer_pi = int(args.pis[i + 2])
                customer_gln = random.choice(tuple(pi_glns[customer_pi - 1]))
                split_num = 1 if (int(split_pi[i]) < 2) else int(split_pi[i])
                for j in range(0, split_num):
                    quantity = int(event['quantity'] / split_num)
                    cte6_data += cte6(event, carrier_gln, factory_gln, customer_gln, quantity)
                    cte6_data[-1]['generator_gln'] = factory_gln
                    cte6_data[-1]['last_pi_gln'] = factory_gln
                    cte6_data[-1]['next_pi_gln'] = carrier_gln
                    if not same_flag:
                        customer_gln = random.choice(tuple(pi_glns[customer_pi - 1]))
            path_data[i].append(cte4_data)
            path_data[i].append(cte5_data)
            path_data[i].append(cte6_data)

        elif int(pis[i]) == 5:
            cte5_data = []
            cte6_data = []
            wholesaler_gln = path_data[i - 1][-1]['customer_gln']

            input_gtins = []
            last_pi_gln = []
            for event in path_data[i - 1]:
                input_gtins.append(event['gtin'])
                last_pi_gln.append(event['generator_gln'])

            if int(merge_gtin[i]) == 0 and int(split_gtin[i]) == 0 :
                last_events = path_data[i-1]
            else:
                cte5_data += cte5_repack(wholesaler_gln, input_gtins)
                cte5_data[-1]['generator_gln'] = wholesaler_gln
                cte5_data[-1]['last_pi_gln'] = last_pi_gln
                cte5_data[-1]['next_pi_gln'] = wholesaler_gln
                last_events = cte5_data
                path_data[i].append(cte5_data)

            for event in last_events:
                carrier_gln = random.choice(tuple(pi_glns[2]))
                customer_pi = int(args.pis[i + 2])
                customer_gln = random.choice(tuple(pi_glns[customer_pi - 1]))
                split_num = 1 if (int(split_pi[i]) < 2) else int(split_pi[i])
                for j in range(0, split_num):
                    try:
                        quantity = int(event['quantity'] / split_num)
                    except KeyError:
                        quantity = int(event['weight'] / split_num)
                    cte6_data += cte6(event, carrier_gln, wholesaler_gln, customer_gln, quantity)
                    cte6_data[-1]['generator_gln'] = wholesaler_gln
                    cte6_data[-1]['last_pi_gln'] = wholesaler_gln
                    cte6_data[-1]['next_pi_gln'] = carrier_gln
                    if not same_flag:
                        customer_gln = random.choice(tuple(pi_glns[customer_pi - 1]))

            path_data[i].append(cte6_data)

        elif int(pis[i]) == 6:
            cte7_data = []
            for j in range(0, len(path_data[i - 1])):
                last_event = path_data[i - 1][j]
                cte7_data += cte7(last_event, 1)
                cte7_data[-1]['generator_gln'] = cte7_data[-1]['retailer_gln']
                cte7_data[-1]['last_pi_gln'] = cte7_data[-1]['generator_gln']
            path_data[i] += cte7_data

    return 0


def main(args):
    pi_num = len(args.pis)
    pi_glns = []
    for i in range(0, 6):
        pi_glns.append( generate_gln_list(3))

    data_dict = {}
    for i in range(0, pi_num):
        data_dict[i] = []

    if args.pis[0] != "1":
        print("Starting point should be 1.")


    merge_num = 1
    merge_point = 0
    for i in range(0, pi_num):
        if int(merge_gtin[i]) >= 1:
            merge_num = int(merge_gtin[i])
            merge_point = int(i)
            break


    for num in range(0, args.sample_num):
        path_data = {}
        for i in range(0, pi_num):
            path_data[i] = []
        for j in range(0, merge_num):
            generate(args.pis[0:merge_point], 0, pi_glns, args.same_pis, path_data)
        generate(args.pis, merge_point, pi_glns, args.same_pis, path_data)

        for i in range(0, pi_num):
            if isinstance(path_data[i][-1], dict):
                if num == 0:
                    data_dict[i].append([])
                data_dict[i][0] += path_data[i]
            else:
                for j in range(0, len(path_data[i])):
                    if num == 0:
                        data_dict[i].append([])
                    for event in path_data[i][j]:
                        data_dict[i][j].append(event)
    # print(data_dict.keys())
    return data_dict


if __name__ == '__main__':
    data_dict = main(args)
    for i in range(0, len(data_dict)):
        for j in range(len(data_dict[i])):
            filename = 'pis-{}-merge_gtin-{}-split_gtin-{}-split_pi-{}-same_pis-{}-pi_index-{}-pi_role-{}-cte-{}.csv'.format(
                args.pis,
                args.merge_gtin,
                args.split_gtin,
                args.split_pi,
                args.same_pis,
                i, args.pis[i], data_dict[i][j][0]['event_type']
            )
            save_to_csv(data_dict[i][j], 'data/' + filename)


