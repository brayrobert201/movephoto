#!/home/bob/.venv/movephoto/bin/python3
import os
import shutil
import calendar
import time
from PIL import UnidentifiedImageError
from PIL import Image

default_watch_dir = "/mnt/c/Users/bob/OneDrive/Pictures/Camera Roll"
robert_watch_dir = "/mnt/c/Users/bob/OneDrive/Pictures/Camera Roll"
# meg_watch_dir = "/mnt/media/nextcloud/meg/files/Photos"
default_destination_dir = "/mnt/c/Users/bob/OneDrive/Camera"

image_extensions = [".jpg", ".jpeg"]
video_extensions = [".mp4", ".mov"]
banned_extensions = [".png"]


def purge_unwanted(watch_dir):
    file_names = os.listdir(watch_dir)
    for file_name in file_names:
        file_ext = os.path.splitext(file_name)[1]
        full_path = f'{watch_dir}/{file_name}'
        if file_ext.lower() in banned_extensions:
            os.remove(full_path)


def move_photos(watch_dir, destination_dir):
    file_names = os.listdir(watch_dir)
    for file_name in file_names:
        file_ext = os.path.splitext(file_name)[1]
        full_path = f'{watch_dir}/{file_name}'
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
                shutil.move(full_path, full_destination)


def move_videos(watch_dir, destination_dir):
    file_names = os.listdir(watch_dir)
    for file_name in file_names:
        file_ext = os.path.splitext(file_name)[1]
        full_path = f'{watch_dir}/{file_name}'
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
                shutil.move(full_path, full_destination)


# purge_unwanted(default_watch_dir)
purge_unwanted(robert_watch_dir)
# move_photos(default_watch_dir, default_destination_dir)
# move_photos(meg_watch_dir, default_destination_dir)
move_photos(robert_watch_dir, default_destination_dir)
# move_videos(default_watch_dir, default_destination_dir)
# move_videos(meg_watch_dir, default_destination_dir)
move_videos(robert_watch_dir, default_destination_dir)
