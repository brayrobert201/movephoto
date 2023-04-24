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


def purge_unwanted(watch_dir, banned_extensions=banned_extensions):
    """
    Remove files with an extension in the banned_extensions list from the
    watch_dir.
    """
    file_names = os.listdir(watch_dir)
    for file_name in file_names:
        # Retrieve only the file extension from the file name
        _, file_ext = os.path.splitext(file_name)
        full_path = os.path.join(watch_dir, file_name)
        # Check if the current file extension is in the banned_extensions list
        if file_ext.lower() in banned_extensions:
            os.remove(full_path)


def move_photos(watch_dir, destination_dir, image_extensions=image_extensions):
    """
    Move the images of the specified extensions inside the watch_dir to the
    corresponding folder inside the destination_dir.
    """
    file_names = os.listdir(watch_dir)
    for file_name in file_names:
        _, file_ext = os.path.splitext(file_name)
        full_path = os.path.join(watch_dir, file_name)
        if file_ext.lower() in image_extensions:
            try:
                # Retrieve the taken date from the image
                image_data = Image.open(full_path)
                date_taken = image_data._getexif()[36867]
                # Split the taken dates into year, month and day
                if type(date_taken) is not list:
                    date_taken = date_taken.split(" ")[0]
                    date_taken = date_taken.split(":")
                year_taken = date_taken[0]
                month_taken = date_taken[1]
                month_name = calendar.month_name[int(month_taken)]
                day_taken = date_taken[2]
                # Create the path to the destination
                full_destination_dir = os.path.join(
                    destination_dir, year_taken, 
                    month_taken + " - " + month_name,
                    year_taken + "-" + month_taken + "-" + day_taken
                )
                full_destination = os.path.join(full_destination_dir, file_name)
                # Create the directory if it does not exist
                if not os.path.exists(full_destination_dir):
                    os.makedirs(full_destination_dir)
                # Move the file if it does not exist
                if not os.path.exists(full_destination):
                    shutil.move(full_path, full_destination)
            except (UnidentifiedImageError, KeyError, TypeError):
                # Skip if any of the handles errors occur
                pass


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
