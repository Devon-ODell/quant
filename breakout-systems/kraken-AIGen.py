import krakenex
from pykrakenapi import KrakenAPI
import pandas as pd
import numpy as np
import time

class KrakenBreakoutSystem:
    def __init__(self, symbol, timeframes=['60', '240', '1440']):
        api = krakenex.API()
        api.load_key('secret.txt')
        self.kraken = KrakenAPI(api)
        self.symbol = symbol
        self.timeframes = timeframes
        self.last_request_time = 0

    def rate_limit(self):
        current_time = time.time()
        if current_time - self.last_request_time < 1:
            time.sleep(1 - (current_time - self.last_request_time))
        self.last_request_time = time.time()

    def fetch_ohlcv_data(self, timeframe):
        self.rate_limit()
        ohlc, _ = self.kraken.get_ohlc_data(self.symbol, interval=int(timeframe))
        return ohlc[['open', 'high', 'low', 'close']]

    def check_breakout(self, df):
        atr = self.calculate_atr(df)
        last_close = df['close'].iloc[-1]
        upper_level = df['high'].rolling(10).max().iloc[-1] + 0.5 * atr
        lower_level = df['low'].rolling(10).min().iloc[-1] - 0.5 * atr
        return last_close > upper_level, last_close < lower_level

    def calculate_atr(self, df):
        tr = np.maximum(
            df['high'] - df['low'],
            np.maximum(
                abs(df['high'] - df['close'].shift()),
                abs(df['low'] - df['close'].shift())
            )
        )
        return tr.rolling(window=14).mean().iloc[-1]

    def run_strategy(self):
        breakouts = [self.check_breakout(self.fetch_ohlcv_data(tf)) for tf in self.timeframes]
        if all(b[0] for b in breakouts):
            self.execute_trade('buy')
        elif all(b[1] for b in breakouts):
            self.execute_trade('sell')
        else:
            print("No confirmed breakout")

    def execute_trade(self, trade_type):
        try:
            self.rate_limit()
            balance = self.kraken.get_account_balance()
            available_balance = float(balance.loc['USDT', 'vol'])
            
            self.rate_limit()
            ticker = self.kraken.get_ticker_information(self.symbol)
            last_price = float(ticker.loc[self.symbol, 'c'][0])
            
            amount = (available_balance * 0.01) / last_price
            
            self.rate_limit()
            order = self.kraken.add_standard_order(
                pair=self.symbol,
                type=trade_type,
                ordertype='market',
                volume=amount,
                validate=False
            )
            print(f"{trade_type.capitalize()} order placed: {order}")
        except Exception as e:
            print(f"Error placing order: {e}")

if __name__ == "__main__":
    system = KrakenBreakoutSystem('XBTUSD')
    system.run_strategy()
