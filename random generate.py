import random
import argparse
import pandas as pd
import math
import numpy as np
import os
from utils import generate_gln_list, cte1, cte2, cte3, cte4, cte5, cte5_repack, cte6, cte7

# 1: fishing company/vessel, 2: auction center; 3: logistic service provider; 4: processing factory;
# 5: distribution center/wholesaler; 6: retailer
pis = "123536"

# 0 : gtin remains the same as before
# n (n>1): number of input gtins at changing point
merge_gtin = "000000"
# 0 : gtin remains the same as before
# n (n>=1): number of output gtins at changing point
split_gtin = "000000"
# n (n>1): number of customers (cte2 or cte6), split path without changing gtin
split_pi = "000200"

parser = argparse.ArgumentParser(description='Data synthesis')
# parser.add_argument('--pis', type=int, default=6, help='Merge paths')
parser.add_argument('--merge_gtin', type=str, default=merge_gtin, help='Control the merge')
parser.add_argument('--split_gtin', type=str, default=split_gtin, help='Control the splitting')
parser.add_argument('--split_pi', type=str, default=split_pi, help='Control the splitting (without changing gtin)')
parser.add_argument('--pis', type=str, default=pis, help='Point of interests in the path')
parser.add_argument('--sample_num', type=int, default=40, help='Number of initial events')
parser.add_argument('--same_pis', type=bool, default=False, help='')
args = parser.parse_args()


def save_to_csv(list, path):
    df = pd.DataFrame(list)
    df.to_csv(path, index=False)


def main(args):
    pi_num = len(args.pis)
    pi_glns = []
    for i in range(0, 6):
        pi_glns.append( generate_gln_list(3))
    data_dict = {}

    pi1_data = []
    if args.pis[0] != "1":
        print("Starting point should be 1.")
    for i in range(0, args.sample_num):
        vessel_gln = random.choice(tuple(pi_glns[0]))
        pi1_data += cte1(vessel_gln, 1)
        pi1_data[-1]['generator_gln'] = vessel_gln


    # print("CTE1 data size: ", len(pi1_data))


    start_index = 1
    if int(args.pis[1]) == 2:
        pi2_data = []
        for i in range(0, len(pi1_data)):
            auction_gln = random.choice(tuple(pi_glns[1]))
            next_pi = int(args.pis[2])
            if next_pi == 3:
                customer_gln = random.choice(tuple(pi_glns[int(args.pis[3]) - 1]))
            else:
                customer_gln = random.choice(tuple(pi_glns[next_pi - 1]))
            carrier_gln = random.choice(tuple(pi_glns[2]))
            pi1_data[i]['generator_gln'] = pi1_data[i]['vessel_gln']
            pi1_data[i]['next_pi_gln'] = auction_gln
            previous_event = pi1_data[i]
            if int(split_pi[1]) < 2:
                pi2_data += cte2(previous_event, auction_gln, customer_gln, previous_event['weight'])
                pi2_data[-1]['generator_gln'] = auction_gln
                pi2_data[-1]['last_pi_gln'] = previous_event['vessel_gln']
                pi2_data[-1]['next_pi_gln'] = carrier_gln
            else:
                for j in range(0, int(split_pi[1])):
                    weight = int(previous_event['weight']/int(split_pi[1]))
                    pi2_data += cte2(previous_event, auction_gln, customer_gln, weight)
                    pi2_data[-1]['generator_gln'] = auction_gln
                    pi2_data[-1]['last_pi_gln'] = previous_event['vessel_gln']
                    pi2_data[-1]['next_pi_gln'] = carrier_gln
                    if not args.same_pis:
                        carrier_gln = random.choice(tuple(pi_glns[2]))
                        if next_pi == 3:
                            customer_gln = random.choice(tuple(pi_glns[int(args.pis[3]) - 1]))
                        else:
                            customer_gln = random.choice(tuple(pi_glns[next_pi - 1]))
        data_dict[0] = [pi1_data]
        data_dict[1] = [pi2_data]
        start_index = 2

    for i in range(start_index, pi_num):
        pi = int(args.pis[i])
        event_data = []
        if pi == 3:
            cte3_data = []
            last_data_list = data_dict[i - 1][-1]
            for j in range(0, len(last_data_list)):
                last_event = last_data_list[j]
                # print(np.array(last_event).shape())
                carrier_gln = last_event['next_pi_gln']
                cte3_data += cte3(last_event, carrier_gln)
                cte3_data[-1]['generator_gln'] = carrier_gln
                cte3_data[-1]['last_pi_gln'] = last_event['generator_gln']
                cte3_data[-1]['next_pi_gln'] = last_event['customer_gln']
            event_data.append(cte3_data)
        elif pi == 6:
            cte7_data = []
            for j in range(0, len(data_dict[i - 1][-1])):
                previous_event = data_dict[i - 1][-1][j]
                cte7_data += cte7(previous_event, 1)
                cte7_data[-1]['generator_gln'] = cte7_data[-1]['retailer_gln']
                cte7_data[-1]['last_pi_gln'] = cte7_data[-1]['generator_gln']
            event_data.append(cte7_data)

        elif pi == 4:
            cte4_data = []
            cte5_data = []
            cte6_data = []
            if int(merge_gtin[i]) == 0 and int(split_gtin[i]) == 0:
                for j in range(0, len(data_dict[i - 1][-1])):
                    previous_event = data_dict[i - 1][-1][j]
                    factory_gln = previous_event['customer_gln']
                    cte4_data += cte4(previous_event['customer_gln'], previous_event['gtin'], 1)
                    cte4_data[-1]['generator_gln'] = factory_gln
                    cte4_data[-1]['last_pi_gln'] = previous_event['generator_gln']
                    cte4_data[-1]['next_pi_gln'] = factory_gln
                event_data.append(cte4_data)
            else:
                input_num = int(merge_gtin[i])
                for j in range(0, int(len(data_dict[i - 1][-1])), input_num):
                    factory_gln = random.choice(tuple(pi_glns[3]))
                    input_gtins = []
                    last_pi_gln = []
                    for k in range(0, input_num):
                        last_pi_gln.append(data_dict[i - 1][-1][j + k]['generator_gln'])
                        data_dict[i - 1][-1][j + k]['customer_gln'] = factory_gln
                        data_dict[i - 2][-1][j + k]['customer_gln'] = factory_gln
                        data_dict[i - 1][-1][j + k]['next_pi_gln'] = factory_gln
                        input_gtins.append(data_dict[i - 1][-1][j + k]['gtin'])
                    cte4_data += cte4(factory_gln, input_gtins, int(split_gtin[i]))
                    cte4_data[-1]['generator_gln'] = factory_gln
                    cte4_data[-1]['last_pi_gln'] = last_pi_gln
                    cte4_data[-1]['next_pi_gln'] = factory_gln

                event_data.append(cte4_data)
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
                if int(split_pi[1]) < 2:
                    cte6_data += cte6(event, carrier_gln, factory_gln, customer_gln, event['quantity'])
                    cte6_data[-1]['generator_gln'] = factory_gln
                    cte6_data[-1]['last_pi_gln'] = factory_gln
                    cte6_data[-1]['next_pi_gln'] = carrier_gln
                else:
                    for j in int(split_pi[i]):
                        quantity = int(event['quantity'] / int(split_pi[i]))
                        cte6_data += cte6(event, carrier_gln, factory_gln, customer_gln, quantity)
                        cte6_data[-1]['generator_gln'] = factory_gln
                        cte6_data[-1]['last_pi_gln'] = factory_gln
                        cte6_data[-1]['next_pi_gln'] = carrier_gln
                        if not args.same_pis:
                            customer_gln = random.choice(tuple(pi_glns[customer_pi - 1]))
            event_data.append(cte5_data)
            event_data.append(cte6_data)

        elif pi == 5:
            cte5_data = []
            cte6_data = []
            if int(merge_gtin[i]) == 1 or int(split_gtin[i]) == 1:
                for j in range(0, len(data_dict[i - 1][-1])):
                    previous_event = data_dict[i - 1][-1][j]
                    wholesaler_gln = previous_event['customer_gln']
                    cte5_data += cte5_repack(wholesaler_gln, previous_event['gtin'])
                    cte5_data[-1]['generator_gln'] = wholesaler_gln
                    cte5_data[-1]['last_pi_gln'] = previous_event['generator_gln']
                    cte5_data[-1]['next_pi_gln'] = wholesaler_gln
                last_events = cte5_data
                event_data.append(cte5_data)
            elif int(merge_gtin[i]) > 1 or int(split_gtin[i]) > 1:
                input_num = int(merge_gtin[i])
                for j in range(0, int(len(data_dict[i - 1][-1])), input_num):
                    previous_event = data_dict[i - 1][-1][j]
                    wholesaler_gln = previous_event['customer_gln']
                    input_gtins = []
                    last_pi_gln = []
                    for k in range(0, input_num):
                        last_pi_gln.append(data_dict[i - 1][-1][j + k]['generator_gln'])
                        data_dict[i - 1][-1][j + k]['customer_gln'] = wholesaler_gln
                        data_dict[i - 2][-1][j + k]['customer_gln'] = wholesaler_gln
                        data_dict[i - 1][-1][j + k]['next_pi_gln'] = wholesaler_gln
                        input_gtins.append(data_dict[i - 1][-1][j + k]['gtin'])
                    cte5_data += cte5_repack(wholesaler_gln, input_gtins)
                    cte5_data[-1]['generator_gln'] = wholesaler_gln
                    cte5_data[-1]['last_pi_gln'] = last_pi_gln
                    cte5_data[-1]['next_pi_gln'] = wholesaler_gln

                event_data.append(cte5_data)
                last_events = cte5_data
            else:
                last_events = data_dict[i - 1][-1]
            for event in last_events:
                carrier_gln = random.choice(tuple(pi_glns[2]))
                customer_pi = int(args.pis[i + 2])
                customer_gln = random.choice(tuple(pi_glns[customer_pi - 1]))
                total_quantity = int(event.get('quantity', event.get('weight', "Lack quantity or weight")))
                if int(split_pi[i]) < 2:
                    cte6_data += cte6(event, carrier_gln, event['next_pi_gln'], customer_gln, total_quantity)
                    cte6_data[-1]['generator_gln'] = event['next_pi_gln']
                    cte6_data[-1]['last_pi_gln'] = event['generator_gln']
                    cte6_data[-1]['next_pi_gln'] = carrier_gln
                else:
                    # print("Split each path to {} sub-paths.".format(split_pi[i]))
                    for j in range(0, int(split_pi[i])):
                        quantity = int(total_quantity / int(split_pi[i]))
                        cte6_data += cte6(event, carrier_gln, event['next_pi_gln'], customer_gln, quantity)
                        cte6_data[-1]['generator_gln'] = event['next_pi_gln']
                        cte6_data[-1]['last_pi_gln'] = event['generator_gln']
                        cte6_data[-1]['next_pi_gln'] = carrier_gln
                        if not args.same_pis:
                            customer_gln = random.choice(tuple(pi_glns[customer_pi - 1]))
            event_data.append(cte6_data)
        data_dict[i] = event_data
    return data_dict


if __name__ == '__main__':
    data_dict = main(args)
    for i in range(0, len(data_dict)):
        for j in range(len(data_dict[i])):
            filename = 'pis-{}-merge_gtin-{}-split_gtin-{}-split_pi-{}-ponit_index-{}-event_index-{}.csv'.format(
                args.pis,
                args.merge_gtin,
                args.split_gtin,
                args.split_pi,
                i, j
            )
            save_to_csv(data_dict[i][j], 'data/' + filename)


