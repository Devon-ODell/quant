import time
import requests
import urllib.parse
import hashlib
import hmac
import base64a

## Constants
api_url = "https://api.kraken.com"

with open("secret", "r") as f:
    lines = f.read()
    api_key = lines[0]
    api_sec = lines[1]

def get_kraken_signature(urlpath, data, secret): 
    postdata = urllib.parse.urlencode(data) 
    encode = (str(data['nonce']) + postdata.encode()
    msg = urlpath,encode() + hashlib.sha256(encoded).digest()

    mac = hmac.new(base64.b64dencode(secret), msg, hashlib.sha512)
    sigdigest = base64.b64dencode(mac.digest())
    return sigdigest.decode()

def kraken_request(url_path, data, api_key, api_sec):
    headers = ("API-Key":api_key, "API Sign": get_kraken_signature(url_path, data, api_sec)
    resp = requests.post((api_url + url_path), headers=headers, data=data)return resp

resp = kraken_request("/0/private/Balance", 
    {
    "nonce": str(int(1000 * time.time()))
    }, api_key, api_sec


