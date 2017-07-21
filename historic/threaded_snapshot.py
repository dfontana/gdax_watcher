""" This is a script for obtaining time series data from GDAX """
import datetime
import os
import shutil
import csv
import threading
from itertools import zip_longest
import gdax

CLIENT = gdax.PublicClient()
GRANULARITY = 1 # second
WAVE_SIZE = 7

def main(srttime=None, endtime=None):
    """
    Breaks down the given time period into digestable request "chunks" that
    the GDAX API can process. Outputs results into a CSV file.
    """
    with open('part_master.csv', 'w') as the_file:
        writer = csv.writer(the_file, dialect='excel')
        writer.writerow(['time', 'low', 'high', 'open', 'close', 'volume'])

    requests = (endtime-srttime).total_seconds() / GRANULARITY

    # Build thread queue
    print("Constructing Threads...")
    threads = define_threads(requests, srttime, endtime)

    # Unleash the threads
    print(str(len(threads)) + " threads constructed.")
    print("Unleashing the Kraken (In waves)...")
    if not os.path.exists("parts"):
        os.makedirs("parts")
    out_file = open("part_master.csv", "a")
    process_threads(out_file, threads)

    # Seal the deal.
    out_file.close()
    print("The Seas Have Settled.")



def process_threads(out_file, threads):
    """
    Iterates over threads in groups of WAVE_SIZE, asynchronously grabbing data,
    writing to a part file, and then merging the parts into the out_file after
    all threads finish.
    """
    wave_index = 1
    wave_size = 0
    for group in grouper(WAVE_SIZE, threads):
        print("\tStarting Wave " + str(wave_index) + "/" + str(len(threads)))
        for thr in group:
            if thr is None:
                continue
            else:
                wave_size += 1
                thr.start()

        for thr in group:
            if thr is None:
                continue
            else:
                thr.join()
        write_parts_to_master(out_file)
        wave_index += 1



def define_threads(requests, srttime, endtime):
    """
    Builds an array of threads to process, where each thread handles a chunk of time
    """
    ths = []
    count = 0
    if requests > 200:
        sframe = srttime
        eframe = sframe + datetime.timedelta(seconds=GRANULARITY*200)
        while eframe <= endtime:
            ths.append(threading.Thread(target=process_frame, args=(sframe, eframe, count)))
            sframe = eframe + datetime.timedelta(seconds=GRANULARITY)
            eframe = sframe + datetime.timedelta(seconds=GRANULARITY*200)
            count += 1
        if eframe > endtime:
            ths.append(threading.Thread(target=process_frame, args=(sframe, endtime, count)))
    else:
        ths.append(threading.Thread(target=process_frame, args=(srttime, endtime, count)))
    return ths



def write_parts_to_master(out_file):
    """
    Writes the current contents of the parts directory to the master csv.
    After which, it deletes the parts & rebuilds folder for the next wave.
    """
    for filename in os.listdir("parts"):
        with open("parts/"+filename) as part:
            for line in part:
                out_file.write(line)
    shutil.rmtree("parts")
    os.makedirs("parts")



def process_frame(start_frame, end_frame, thread_count):
    """
    Makes a call to the historic endpoint for the given time period, writing results
    to file. Sometimes the API returns "message" - that data row is filtered out.
    Additionally, the timestamp is in epoch time - which has been converted to
    human readable output in UTC time.
    """
    subarray = CLIENT.get_product_historic_rates('ETH-USD', start=start_frame,
                                                 end=end_frame, granularity=GRANULARITY)

    with open('parts/part_'+str(thread_count)+'.csv', 'w') as the_file:
        writer = csv.writer(the_file, dialect='excel')
        for row in subarray:
            if row[0] == 'm':
                break
            row[0] = datetime.datetime.fromtimestamp(row[0]).strftime('%x %X')
            writer.writerow(row)



def grouper(chunk_size, iterable, fillvalue=None):
    """
    Splits an array into chunk_sized subarrays, filling in empty spaces
    with None by default.
    """
    args = [iter(iterable)] * chunk_size
    return zip_longest(fillvalue=fillvalue, *args)



START = datetime.datetime(2017, 1, 1, 6, 0)
END = datetime.datetime.now()
main(START, END)
