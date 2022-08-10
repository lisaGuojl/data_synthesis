import random
import string
import pandas as pd
from time import gmtime, strftime
import datetime
import uuid
import json

P1Location = {"Denpasar": "-8.650000,115.216667"}
P2Location = {"Banyuwangi": "-8.2192335,114.36922670000001"}
P3Location = {"Probolinggo": "-7.756928,113.211502", "Malang": "-7.983908, 112.621391"}
P4Location = {"Surabaya": "-7.250445, 112.768845", "Kalten": "-7.703403, 110.600502"}
# P5Location = ["-7.3262096,110.461207", "-7.430185,109.1626644"]
P6Location = {"Semarang": "-6.966667, 110.416664", "Jakarta": "-6.200000, 106.816666"}
species = ["Salmon", "Cod", "Tuna"]

# generate serial number or batch\lot number 2 numeric digits + up to 20 characters
def generate_serialno():
    length = random.randint(1, 21)
    n = ''.join([str(random.randint(0, 10)) for _ in range(2)])
    x = ''.join(random.choice(string.ascii_uppercase) for _ in range(length))
    return n + x


def generate_gln_list(num=1):
    glns = []
    for i in range(0, num):
        glns.append(generate_gln())
    return glns


def generate_gtin():
    return str(''.join([str(random.randint(0, 9)) for _ in range(14)]))


def generate_sscc():
    return str('0' + ''.join([str(random.randint(0, 9)) for _ in range(17)]))


def generate_gln():
    return str(''.join([str(random.randint(0, 9)) for _ in range(13)]))


def generate_string(length):
    return str(''.join(random.choice(string.ascii_lowercase + string.ascii_uppercase) for _ in range(length)))



def cte1(vessal_gln, number=1):
    # vessal_gln = generate_gtin()

    data = []
    city, loc = random.choice(list(P1Location.items()))
    for i in range(0, number):
        cte_data = {
            'previous_key': str(uuid.uuid4()),
            'event_id': str(uuid.uuid4()),
            'event_type': 1,
            'vessel_gln': vessal_gln,
            'gtin': generate_gtin(),
            'serial_number': generate_serialno(),
            'weight': random.randint(500, 10000),
            'event_time': strftime("%Y-%b-%dT%H:%M:%S +0000", gmtime()),
            'location_name': city,
            'location_coordinate': loc,
            'catch_date': strftime("%Y-%b-%d", gmtime()),
            'vessal_owner_name': generate_string(7),
            'species': random.choice(tuple(species)),
            'economic_zone': generate_gln(),
            'first_freeze_date': strftime("%Y-%b-%d", gmtime()),
            'catch_certificate_id': generate_gtin(),
            'conservation_reference_size': str(random.randint(10, 50)) + 'cm',
            'catch_area': generate_gln(),
            'company_name': generate_string(7),
            'generator_type': "fishing vessel"
        }
        cte_data['new_key'] = cte_data['previous_key']
        data.append(cte_data)

    return data


def cte2(event, auction_gln, customer_gln, weight):
    data = []
    city, loc = random.choice(list(P1Location.items()))
    if event['weight'] > 0:
        if event['weight'] - weight <= 0:
            weight = event['weight']
        # event['weight'] -= weight
        cte_data = {
            'previous_key': event['previous_key'],
            'event_id': str(uuid.uuid4()),
            'event_type': 2,
            'auction_gln': auction_gln,
            'supplier_gln': event['vessel_gln'],
            'customer_gln': customer_gln,
            'gtin': event['gtin'],
            'serial_number': event['serial_number'],
            'weight': weight,
            'event_time': str((datetime.datetime.today() + datetime.timedelta(days=1)).strftime('%Y-%b-%dT%H:%M:%S +0000')),
            'location_name': city,
            'location_coordinate': loc,
            'product_name': generate_string(5),
            'comapny_name': generate_string(7),
            'generator_type': "auction center"
        }
        cte_data['new_key'] = cte_data['previous_key']
        data.append(cte_data)

    return data


def cte3(event, carrier_gln):
    try:
        weight = event['quantity']
    except KeyError:
        weight = event['weight']
    base = datetime.datetime.strptime(event['event_time'], '%Y-%b-%dT%H:%M:%S +0000')
    data = [{
        'previous_key': event['previous_key'],
        'new_key': event['previous_key'],
        'event_id': str(uuid.uuid4()),
        'event_type': 3,
        'supplier_gln': event['supplier_gln'],
        'customer_gln': event['customer_gln'],
        'carrier_gln': carrier_gln,
        'sscc': generate_gln(),
        'gtin': event['gtin'],
        'serial_number': event['serial_number'],
        'weight': weight,
        'event_time': str((base + datetime.timedelta(days=1)).strftime('%Y-%b-%dT%H:%M:%S +0000')),
        'location_name': "",
        'location_coordinate': "",
        'departure_gln': generate_gln(),
        'temperature': float("{0:.1f}".format(random.uniform(-10, 0))),
        'comapny_name': generate_string(7),
        'generator_type': "logistic services provider"
    }]

    return data


def cte4(factory_gln, input_gtins, output_number, previous_keys):
    base = datetime.datetime.today() + datetime.timedelta(days=4)

    # randomly generate 1 or 2 output_gtins
    output_gtins = []
    for i in range(0, output_number):
        output_gtins.append(generate_gtin())
    city, loc = random.choice(list(P2Location.items()))

    data = [{
        'event_id': str(uuid.uuid4()),
        'event_type': 4,
        'factory_gln': factory_gln,
        'input_gtin': input_gtins,
        'output_gtin': output_gtins,
        'serial_number': generate_serialno(),
        'quantity': str(random.randint(100, 500)),
        'event_time': str((base + datetime.timedelta(days=1)).strftime('%Y-%b-%dT%H:%M:%S +0000')),
        'location_name':city,
        'location_coordinate': loc,
        'brand_name': generate_string(10),
        # 'product_method': generate_string(3),
        # 'ingredient_statement': generate_string(20),
        'storage_state': 'PREVIOUSLY_FROZEN',
        'expiration_date': str((base + datetime.timedelta(days=60)).strftime('%Y-%m-%d')),
        'comapny_name': generate_string(7),
        'generator_type': "processing company"
    }]
    if len(previous_keys) > 1:
        data[0]['new_key'] = str(uuid.uuid4())
        data[0]['previous_key'] = previous_keys
    else:
        data[0]['new_key'] = previous_keys[0]
        data[0]['previous_key'] = previous_keys[0]

    return data


def cte5(event, generator_gln):
    data = []
    base = datetime.datetime.strptime(event['event_time'], '%Y-%b-%dT%H:%M:%S +0000')
    # pack each processing output separately
    for input_gtin in event['output_gtin']:
        cte5_data = {
            'previous_key': event['previous_key'],
            'event_id': str(uuid.uuid4()),
            'event_type': 5,
            'generator_gln': generator_gln,
            'input_gtin': input_gtin,
            'output_gtin': generate_gtin(),
            'serial_number': generate_serialno(),
            'quantity': int((random.randint(100, 500)) / 4),
            'event_time': str((base + datetime.timedelta(days=1)).strftime('%Y-%b-%dT%H:%M:%S +0000')),
            'location_name': event['location_name'],
            'location_coordinate': event['location_coordinate'],
            'net_contain': int((random.randint(0, 9))),
            'packing_type_code': generate_string(3),
            'packing_material': 'PLASTIC_THERMOPLASTICS',
            'recycling_process_type': 'Recyclable',
            'comapny_name': generate_string(7),
            'generator_type': "processing company"

        }
        if len(event['output_gtin']) > 1:
            cte5_data['new_key'] = str(uuid.uuid4())
        else:
            cte5_data['new_key'] = event['previous_key'][0]


        data.append(cte5_data)

    return data


def cte5_repack(generator_gln, input_gtin, previous_key, output_num=1):
    data = []
    # base = datetime.datetime.strptime(event['event_time'], '%Y-%b-%dT%H:%M:%S +0000')
    # pack each processing output separately
    output_gtin = []
    city, loc = random.choice(list(P4Location.items()))
    for i in range(0, output_num):
        output_gtin.append(generate_gtin())
    cte5_data = {
        'previous_key': previous_key,
        'new_key': previous_key,
        'event_id': str(uuid.uuid4()),
        'event_type': 5,
        'generator_gln': generator_gln,
        'input_gtin': input_gtin,
        'output_gtin': output_gtin,
        'serial_number': generate_serialno(),
        'quantity': int((random.randint(100, 500)) / 4),
        'event_time': str((datetime.datetime.today() + datetime.timedelta(days=1)).strftime('%Y-%b-%dT%H:%M:%S +0000')),
        'location_name': city,
        'location_coordinate': loc,
        'net_contain': int((random.randint(0, 9))),
        'packing_type_code': generate_string(3),
        'packing_material': 'PLASTIC_THERMOPLASTICS',
        'recycling_process_type': 'Recyclable',
        'comapny_name': generate_string(7),
        'generator_type': "warehouse"
    }
    data.append(cte5_data)

    return data


def cte6(event, carrier_gln, supplier_gln, customer_gln, quantity, new_key, city="", loc=""):
    try:
        gtin = event['output_gtin']
    except KeyError:
        gtin = event['gtin']
    data = []
    base = datetime.datetime.strptime(event['event_time'], '%Y-%b-%dT%H:%M:%S +0000')
    cte6_data = {
        'previous_key': event['previous_key'],
        'new_key': new_key,
        'event_id': str(uuid.uuid4()),
        'event_type': 6,
        'supplier_gln': supplier_gln,
        'customer_gln': customer_gln,
        'carrier_gln': carrier_gln,
        'sscc': generate_gln(),
        'gtin': gtin,
        'serial_number': event['serial_number'],
        'quantity': quantity,
        'event_time': str((base + datetime.timedelta(days=1)).strftime('%Y-%b-%dT%H:%M:%S +0000')),
        'location_name': city,
        'location_coordinate': loc,
        'departure_gln': generate_gln(),
        'weight': float("{0:.1f}".format(random.uniform(10, 100))),
        'temperature': float("{0:.1f}".format(random.uniform(-10, 0))),
        'comapny_name': generate_string(7),
        'generator_type': ""
    }
    data.append(cte6_data)

    return data


def cte7(event, quantity):
    data = []
    base = datetime.datetime.strptime(event['event_time'], '%Y-%b-%dT%H:%M:%S +0000')
    city, loc = random.choice(list(P4Location.items()))
    cte_data = {
        'previous_key': event['previous_key'],
        'new_key': event['previous_key'],
        'event_id': str(uuid.uuid4()),
        'event_type': 7,
        'retailer_gln': event['customer_gln'],
        'gtin': event['gtin'],
        'serial_number': event['serial_number'],
        'quantity': quantity,
        'event_time': str((base + datetime.timedelta(days=1)).strftime('%Y-%b-%dT%H:%M:%S +0000')),
        'location_name': city,
        'location_coordinate': loc,
        'price': float("{0:.2f}".format(random.uniform(3, 50))),
        'company_name': generate_string(7),
        'generator_type': "retailer"
    }
    data.append(cte_data)

    return data

# def cte7(event, total_quantity):
#     data = []
#     for i in range(0, 10):
#         quantity = int(random.randint(1, 10))
#         if total_quantity - quantity <= 0:
#             quantity = total_quantity
#         total_quantity -= quantity
#         cte_data = {
#             'event_type': 7,
#             'retailer_gln': event['customer_gln'],
#             'gtin': event['gtin'],
#             'serial_number': event['serial_number'],
#             'quantity': quantity,
#             'event_time': strftime("%Y-%b-%dT%H:%M:%S +0000", gmtime()),
#             'location_gln': generate_gln(),
#             'price': float("{0:.2f}".format(random.uniform(3, 50)))
#         }
#         data.append(cte_data)
#
#     return data


if __name__ == '__main__':
    # for i in range(0,10):
    #     print(generate_gtin())
    time = str(strftime("%Y-%b-%dT%H:%M:%S +0000", gmtime()))
    t2 = datetime.datetime.strptime(time, '%Y-%b-%dT%H:%M:%S +0000')
    print(time, t2)