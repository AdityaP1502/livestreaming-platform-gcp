import json
import tkinter as tk
from tkinter import filedialog
from tkinter import messagebox
from tkinter.filedialog import askopenfilename
from tkinter import *
import subprocess
import datetime
import requests

global formatted_datetime
current_datetime = datetime.datetime.now()
formatted_datetime = current_datetime.strftime("%Y-%m-%d %H-%M-%S")

# Function to start streaming
def start_streaming(username):

    user_data = {'username' : username}
    json_login = json.dumps(user_data)
    login_headers = {'Content-Type': 'application/json'}

    api_url = "http://34.101.36.32:8000/stream"

    response = requests.post(api_url, data=json_login, headers=login_headers)

    if response.status_code == 201: 
        
        stream_data = response.json().get('data', {})
        global stream_id
        global bucket_url
        global rtsp_url
        stream_id = stream_data.get('stream', {}).get('stream-id')
        rtsp_url = stream_data.get('stream', {}).get('rtsp-url')
        bucket_url = stream_data.get('BucketUrl')
        print(stream_id)
        print(rtsp_url)
        print(bucket_url)

        create_description(username)

    else:
        print(f"Failed to start streaming. API returned status code {response.status_code}")

# Function to stop streaming
def stop_streaming(username):

    api_url = f"http://34.101.36.32:8000/stream/{username}/{stream_id}"

    response = requests.delete(api_url)

    if response.status_code == 201:
        update_log(f"Stream stopped")
    else:
         print(f"Failed to stop streaming. API returned status code {response.status_code}")

    p.kill()

# Function to upload picture
def upload_pict():
    global pict_path
    pict_path = filedialog.askopenfilename(title="Select picture", filetypes=[("Image files", "*.png;*.jpg;*.jpeg;*.gif")])

    if pict_path:
        update_log(f"Selected picture path: {pict_path}")
        browse.delete(1.0, tk.END)  # Clear previous content
        browse.insert(tk.END, pict_path)
    else:
        update_log("Image selection canceled.")
        browse.delete(1.0, tk.END)  # Clear previous content
    

# Function to create and display dashboard
def create_dashboard(username):
    global dashboard_window
    dashboard_window = tk.Toplevel()
    dashboard_window.title("Dashboard")
    dashboard_window.geometry("1000x500")

    global log
    log = tk.Text(master=dashboard_window)
    log.place(x=300, y=100)
    log.config(width=83, height=25)
    log.configure(state='disabled')

    # Elements in the dashboard
    font_title = ("montserrat", 24, "bold")
    tk.Label(dashboard_window, text="Live-Streaming Client", font=font_title).place(x=250, y=10)

    font = ("montserrat", 16, "bold")
    tk.Button(dashboard_window, text="Start", command=lambda: start_streaming(username), height=2, width=8, bg='#32a871', fg='black',
              font=font).place(x=50, y=100)

    tk.Button(dashboard_window, text="Stop", command=lambda: stop_streaming(username), height=2, width=8, bg='red', fg='black',
              font=font).place(x=50, y=200)

    font2 = ("montserrat", 12, "bold")
    
    tk.Button(dashboard_window, text="❌", command=lambda: on_closing(username), height=2, width=4, bg='red', fg='black', font=font2).place(x=920, y=10)

def create_description(username):
    global desc_window, pict_path, browse
    desc_window = tk.Toplevel()
    desc_window.title("Description")
    desc_window.geometry("600x300")

    font2 = ("montserrat", 12, "bold")

    title_label = tk.Label(desc_window, text="Title:", font=font2)
    title_label.place(x=10,y=5)

    title_entry = tk.Entry(desc_window)
    title_entry.place(x=10,y=30)
    title_entry.configure(width=95)

    thumbnail_label = tk.Label(desc_window, text="Thumbnail file:", font=font2)
    thumbnail_label.place(x=10,y=60)

    browse = tk.Text(master=desc_window)
    browse.place(x=10,y=85)
    browse.config(width=70, height=2)

    tk.Button(desc_window, text="Browse", command=upload_pict,font=font2).place(x=10,y=130)
    
    tk.Button(desc_window, text="Upload", command=lambda: upload_desc(username, title_entry), height=2, width=14,
              font=font2, bg='#32a871').place(x=10,y=220)
    
    

def upload_desc(username, title_entry):
    user_title = {'name' : title_entry.get(),
                 'created-at' : formatted_datetime }
    json_title = json.dumps(user_title)
    title_headers = {'Content-Type': 'application/json'}

    api_url = f"http://34.101.36.32:8000/stream/{username}/{stream_id}/metadata"

    pict_upload = [f'curl', '-X', 'PUT','-H', 'Content-Type: image/png', '--upload-file', pict_path, bucket_url]
    subprocess.Popen(pict_upload)

    response = requests.post(api_url, data=json_title, headers=title_headers)

    if response.status_code == 201:
        update_log('Title saved')
        update_log('Thumbnail uploaded')
        desc_window.withdraw()

        ffmpeg_command = [
        'ffmpeg', '-video_size', '1920x1080', '-f', 'gdigrab', '-re', '-i', 'desktop',
        '-vcodec', 'libx264', '-tune', 'zerolatency', '-preset', 'ultrafast',
        '-f', 'rtsp', rtsp_url]

        global p
        p = subprocess.Popen(ffmpeg_command)
        update_log("Stream started")  
        #update_log(f"Streaming at ")

    else:
        messagebox.showinfo('Upload failed!')

# Function to update log
def update_log(message):
    global log
    log.configure(state='normal')
    log.insert(tk.END, "\n")
    log.insert(tk.END, datetime.datetime.now())
    log.insert(tk.END, "\n")
    log.insert(tk.END, message)
    log.insert(tk.END, "\n")
    log.configure(state='disabled')

# Function to open dashboard after successful login
def open_dashboard(username):
    login_window.withdraw()  # Hide login window
    create_dashboard(username)

# Function to check login credentials
def check_login(username, password):
    
    login_data = {'username' : username, 
                  'password' : password}
    json_login = json.dumps(login_data)
    login_headers = {'Content-Type': 'application/json'}

    api_url = "http://34.101.36.32:8000/login"

    response = requests.post(api_url, data=json_login, headers=login_headers)

    if response.status_code == 201:
        open_dashboard(username)  
        update_log("Login success")  

    else:
        print(f"Failed to start streaming. API returned status code {response.status_code}")
        messagebox.showerror("Login Error", "Wrong username or password!")


# Function to check signup credentials
def check_signup(username, password):
    existing_username = "admin"
    return username != existing_username

# Function to handle login button click
def login():
    entered_username = username_entry.get()
    entered_password = password_entry.get()

    check_login(entered_username, entered_password)

# Function to handle signup button click
def signup():
    entered_username = username_entry.get()
    entered_password = password_entry.get()

    if check_signup(entered_username, entered_password):
        messagebox.showinfo("Sign Up", "Sign up successful!")
    else:
        messagebox.showerror("Sign Up Error", "Username already exists")

def on_closing(username):
    stop_streaming(username)  # Hentikan proses streaming sebelum menutup aplikasi
    login_window.destroy()
    dashboard_window.destroy()

# Create login window
login_window = tk.Tk()
login_window.title("Login Page")
login_window.geometry("200x250")
login_window.protocol("WM_DELETE_WINDOW", on_closing)

# Elements in the login window
username_label = tk.Label(login_window, text="Username:")
username_label.pack(pady=10)

username_entry = tk.Entry(login_window)
username_entry.pack(pady=10)

password_label = tk.Label(login_window, text="Password:")
password_label.pack(pady=10)

password_entry = tk.Entry(login_window, show="*")
password_entry.pack(pady=10)

login_button = tk.Button(login_window, text="Login", command=login)
login_button.pack(pady=10)

signup_button = tk.Button(login_window, text="Sign Up", command=signup)
signup_button.pack(pady=10)


# Run the main loop
login_window.mainloop()
