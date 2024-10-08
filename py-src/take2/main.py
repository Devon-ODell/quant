# # # Fetches the 2023 historical data for a specified Crypto portfolio, 
# # # runs a backtest composed of a series of breakout strategies,
# # # and visualizes the efficacy of each breakout strat, in Python
# # # Authored By: Devon ODell

import os
import pandas as pd
import numpy as np
from typing import List, Dict
from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build 
from googleapiclient.errors import HttpError



# Constants #

SCOPES = ["https://googleapis.com/auth/spreadsheets"]
SPREADSHEET_ID = "1lhZl4paLrBRaVPbduRG4IvAPlT7ODZgW3c24NLi_M5I"
SAMPLE_RANGE_NAME = "Class Data!A2:E"
CRYPTO_SYMBOLS = ['BTC', 'ADA', 'LINK', 'ETH', 'LTC', 'XMR', 'XRP', 'SOL', 'UNI', 'AAVE']


class Backtester:
    def __init__( data: Dict[str, pd.DataFrame], initial_capital: float):
        self.data = data
        self.capital = initial_capital
        self.positions: Dict[str, int] = {}
        self.trades: List[Dict] = []
    
    def run(self, strategy):
        # Implement backtesting logic here
        pass
    
    def calculate_returns(self):
        # Calculate and return performance metrics
        pass

def example_strategy(data: pd.DataFrame) -> str:
    # Implement your trading strategy here
    # Return 'buy', 'sell', or 'hold'
    pass

backtester = Backtester(data, initial_capital=100000)
results = backtester.run(example_strategy)
performance = backtester.calculate_returns()

def get_google_sheets_credentials():
    creds = None
    if os.path.exists("token.json"):
        creds = Credentials.from_authorized_user_file("token.json", SCOPES)
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            flow = InstalledAppFlow.from_client_secrets_file(
                "credentials.json", SCOPES
            )
            creds = flow.run_local_server(port=0)
        with open("token.json", "w") as token:
            token.write(creds.to_json())
    return creds

def fetch_google_sheets_data(creds):
    try:
        service = build("sheets", "v4", credentials=creds)
        sheet = service.spreadsheets()
        result = sheet.values().get(spreadsheetId=SPREADSHEET_ID, range=SAMPLE_RANGE_NAME).execute()
        return result.get("values", [])
    except HttpError as err:
        print(f"An error occurred: {err}")
        return None

def process_sheet_data(values):
    if not values:
        print("No data found.")
        return

    print("Name, Major:")
    for row in values:
        # Print columns A and E, which correspond to indices 0 and 4.
        print(f"{row[0]}, {row[4]}")

def fetch_crypto_data():
   # Implement function to fetch historical data for CRYPTO_SYMBOLS
    # This should return a Dict[str, pd.DataFrame]
   

    try:
        service = build("sheets", "v4", credentials=creds)
        sheet = service.spreadsheets()
        
        data = {}
        for symbol in CRYPTO_SYMBOLS:
            range_name = f"{symbol}!A:G"  # Adjust if your data is in different columns
            result = sheet.values().get(spreadsheetId=SPREADSHEET_ID, range=range_name).execute()
            values = result.get('values', [])
            
            if not values:
                print(f'No data found for {symbol}.')
                continue
            
            df = pd.DataFrame(values[1:], columns=values[0])
            df['Date'] = pd.to_datetime(df['Date'])
            df.set_index('Date', inplace=True)
            for col in ['Price', 'Open', 'High', 'Low', 'Vol.']:
                df[col] = pd.to_numeric(df[col].str.replace(',', ''), errors='coerce')
            
            data[symbol] = df
        
        return data
    except HttpError as err:
        print(f"An error occurred: {err}")
        return None 
    pass 

def main():
    def main():
    creds = get_google_sheets_credentials()
    crypto_data = fetch_crypto_data(creds)
    
    if crypto_data:
        backtester = Backtester(crypto_data, initial_capital=100000)
        results = backtester.run(example_strategy)
        performance = backtester.calculate_returns()
        
        # Convert performance metrics to a DataFrame
        results_df = pd.DataFrame(performance)
        
        # Write results back to Google Sheets
        write_backtest_results(creds, results_df)

if __name__ == "__main__":
    main()
