#!/usr/bin/python3

import os
import shutil
import sys
import calendar
import time
from PIL import UnidentifiedImageError
from PIL import Image
import time
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

default_watch_dir = "/home/bob/Dropbox/Camera Uploads"
robert_watch_dir = "/mnt/media/nextcloud/brayrobert201/files/InstantUpload/Camera"
meg_watch_dir = "/mnt/media/nextcloud/meg/files/Photos"
default_destination_dir = "/mnt/media/Syncthing/Camera"

directories = [robert_watch_dir, meg_watch_dir]
#directories = ['/tmp', '/home/bob']

observers = []

image_extensions = [".jpg", ".jpeg"]
video_extensions = [".mp4", ".mov"]
banned_extensions = [".png"]

class Watcher:

    def __init__(self):
        self.observer = Observer()

    def run(self):
        event_handler = Handler()
        #self.observer.schedule(event_handler, self.DIRECTORY_TO_WATCH, recursive=True)
        for directory in directories:
            self.observer.schedule(event_handler, directory, recursive=True)
            observers.append(self.observer)
        self.observer.start()
        try:
            while True:
                time.sleep(5)
        except:
            self.observer.stop()
            print("Error")

        self.observer.join()


class Handler(FileSystemEventHandler):

    @staticmethod
    def on_any_event(event):
        if event.is_directory:
            return None

        elif event.event_type == 'modified':
            file_ext = os.path.splitext(event.src_path,)[1]
            if file_ext.lower() in image_extensions:
                print(f"Copied {event.src_path}")
                move_single_photo(event.src_path, default_destination_dir)
            elif file_ext.lower() in video_extensions:
                print(f"Copied {event.src_path}")
                move_single_video(event.src_path, default_destination_dir)
            elif file_ext.lower() in banned_extensions:
                print(f"Removed {event.src_path}")
                os.remove(event.src_path)

def move_single_photo(full_path, destination_dir):
    file_name = os.path.basename(full_path)
    file_ext = os.path.splitext(file_name)[1]
    if file_ext.lower() in image_extensions:
        try:
            image_data = Image.open(full_path)
        except UnidentifiedImageError:
            print(f'Cant identify {full_path}')
            pass
        try:
            date_taken = image_data._getexif()[36867]
        except KeyError:
            pass
        except TypeError:
            pass
        if type(date_taken) is not list:
            date_taken = date_taken.split(" ")[0]
            date_taken = date_taken.split(":")
        year_taken = date_taken[0]
        month_taken = date_taken[1]
        month_name = calendar.month_name[int(month_taken)]
        day_taken = date_taken[2]
        full_destination_dir = f'{destination_dir}/{year_taken}/{month_taken} - {month_name}/{year_taken}-{month_taken}-{day_taken}'
        full_destination = f'{full_destination_dir}/{file_name}'
        if not os.path.exists(full_destination_dir):
            os.makedirs(full_destination_dir)
        if not os.path.exists(full_destination):
            shutil.copy(full_path, full_destination)

def move_single_video(full_path, destination_dir):
    file_name = os.path.basename(full_path)
    file_ext = os.path.splitext(file_name)[1]
    if file_ext.lower() in video_extensions:
        date_taken = os.stat(full_path).st_mtime
        date_taken = time.strftime('%Y-%m-%d', time.localtime(date_taken))
        date_taken = date_taken.split("-")
        if type(date_taken) is not list:
            date_taken = date_taken.split(" ")[0]
            date_taken = date_taken.split(":")
        year_taken = date_taken[0]
        month_taken = date_taken[1]
        month_name = calendar.month_name[int(month_taken)]
        day_taken = date_taken[2]
        full_destination_dir = f'{destination_dir}/{year_taken}/{month_taken} - {month_name}/{year_taken}-{month_taken}-{day_taken}'
        full_destination = f'{full_destination_dir}/{file_name}'
        if not os.path.exists(full_destination_dir):
            os.makedirs(full_destination_dir)
        if not os.path.exists(full_destination):
            shutil.copy(full_path, full_destination)


if __name__ == '__main__':
    w = Watcher()
    w.run()
