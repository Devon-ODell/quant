
import krakenex
import pandas as pd
import numpy as np
import time
import logging
from datetime import datetime
from pykrakenapi import KrakenAPI

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class TurtleTrader:
    def __init__(self, api_key, api_sec):
        self.kraken = krakenex.API(api_key, api_sec)
        self.k = KrakenAPI(self.kraken)
        
        self.pair_btc = 'XXBTZUSD'
        self.pair_xmr = 'XXMRZUSD'
        self.timeframe = 240  # 4 hour candles
        
        # Strategy parameters
        self.sma_period = 200
        self.donchian_period = 20
        self.atr_period = 14
        self.risk_percentage = 0.03  # 3% risk per trade
        
    def get_candles(self, pair, since=None):
        try:
            ohlc, last = self.k.get_ohlc_data(pair, interval=self.timeframe, since=since)
            return ohlc
        except Exception as e:
            logger.error(f"Error fetching candles: {e}")
            return None

    def calculate_sma(self, data, period):
        return data['close'].rolling(window=period).mean()

    def calculate_donchian(self, data, period):
        upper = data['high'].rolling(window=period).max()
        lower = data['low'].rolling(window=period).min()
        return upper, lower

    def calculate_atr(self, data, period):
        high = data['high']
        low = data['low']
        close = data['close']
        
        tr1 = high - low
        tr2 = abs(high - close.shift(1))
        tr3 = abs(low - close.shift(1))
        
        tr = pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)
        atr = tr.rolling(window=period).mean()
        
        return atr

    def calculate_adx(self, data, period=14):
        high = data['high']
        low = data['low']
        close = data['close']
        
        # Calculate +DM and -DM
        high_diff = high - high.shift(1)
        low_diff = low.shift(1) - low
        
        plus_dm = pd.Series(0.0, index=high_diff.index)
        minus_dm = pd.Series(0.0, index=low_diff.index)
        
        plus_dm[((high_diff > 0) & (high_diff > low_diff))] = high_diff
        minus_dm[((low_diff > 0) & (low_diff > high_diff))] = low_diff
        
        # Calculate TR
        tr = self.calculate_atr(data, 1)
        
        # Calculate smoothed +DM, -DM, and TR
        smoothed_plus_dm = plus_dm.rolling(window=period).mean()
        smoothed_minus_dm = minus_dm.rolling(window=period).mean()
        smoothed_tr = tr.rolling(window=period).mean()
        
        # Calculate +DI and -DI
        plus_di = 100 * (smoothed_plus_dm / smoothed_tr)
        minus_di = 100 * (smoothed_minus_dm / smoothed_tr)
        
        # Calculate DX and ADX
        dx = 100 * abs(plus_di - minus_di) / (plus_di + minus_di)
        adx = dx.rolling(window=period).mean()
        
        return adx

    def calculate_chop(self, data, period=14):
        atr = self.calculate_atr(data, 1)
        highest_high = data['high'].rolling(window=period).max()
        lowest_low = data['low'].rolling(window=period).min()
        
        chop = 100 * np.log10(atr.rolling(window=period).sum() / (highest_high - lowest_low)) / np.log10(period)
        return chop

    def calculate_indicators(self, candles):
        try:
            # Calculate all indicators
            sma = self.calculate_sma(candles, self.sma_period)
            upper_band, lower_band = self.calculate_donchian(candles, self.donchian_period)
            adx = self.calculate_adx(candles)
            chop = self.calculate_chop(candles)
            atr = self.calculate_atr(candles, self.atr_period)
            
            return {
                'sma': sma.iloc[-1],
                'donchian_upper': upper_band.iloc[-1],
                'donchian_lower': lower_band.iloc[-1],
                'adx': adx.iloc[-1],
                'chop': chop.iloc[-1],
                'atr': atr.iloc[-1]
            }
        except Exception as e:
            logger.error(f"Error calculating indicators: {e}")
            return None

    def place_order(self, pair, side, qty, price, stop_loss):
        try:
            # Place the main order
            order_params = {
                'pair': pair,
                'type': side,
                'ordertype': 'limit',
                'price': str(price),
                'volume': str(qty)
            }
            
            main_order = self.k.add_standard_order(**order_params)
            
            # Place stop loss order
            if main_order:
                sl_params = {
                    'pair': pair,
                    'type': 'sell' if side == 'buy' else 'buy',
                    'ordertype': 'stop-loss',
                    'price': str(stop_loss),
                    'volume': str(qty)
                }
                self.k.add_standard_order(**sl_params)
            
            return main_order
        except Exception as e:
            logger.error(f"Error placing order: {e}")
            return None

    def get_position_size(self, current_price, stop_loss, pair):
        try:
            balance = float(self.k.get_account_balance()['tb'])
            risk_amount = balance * self.risk_percentage
            position_size = risk_amount / abs(float(current_price) - float(stop_loss))
            return position_size
        except Exception as e:
            logger.error(f"Error calculating position size: {e}")
            return None

    def check_signals(self, pair, indicators, current_price):
        if indicators is None:
            return

        # Long signal
        if (current_price > indicators['donchian_upper'] and 
            current_price > indicators['sma'] and 
            indicators['adx'] > 30 and 
            indicators['chop'] < 40):
            
            stop_loss = current_price - (indicators['atr'] * 2.5)
            position_size = self.get_position_size(current_price, stop_loss, pair)
            
            if position_size:
                return self.place_order(pair, 'buy', position_size, current_price, stop_loss)

        # Short signal
        elif (current_price < indicators['donchian_lower'] and 
              current_price < indicators['sma'] and 
              indicators['adx'] > 30 and 
              indicators['chop'] < 40):
            
            stop_loss = current_price + (indicators['atr'] * 2.5)
            position_size = self.get_position_size(current_price, stop_loss, pair)
            
            if position_size:
                return self.place_order(pair, 'sell', position_size, current_price, stop_loss)

        return None

    def run_strategy(self):
        while True:
            try:
                # Process BTC pair
                btc_candles = self.get_candles(self.pair_btc)
                if btc_candles is not None:
                    btc_indicators = self.calculate_indicators(btc_candles)
                    current_price = float(self.k.get_ticker_information(self.pair_btc)['c'][0][0])
                    self.check_signals(self.pair_btc, btc_indicators, current_price)

                # Process XMR pair
                xmr_candles = self.get_candles(self.pair_xmr)
                if xmr_candles is not None:
                    xmr_indicators = self.calculate_indicators(xmr_candles)
                    current_price = float(self.k.get_ticker_information(self.pair_xmr)['c'][0][0])
                    self.check_signals(self.pair_xmr, xmr_indicators, current_price)

                # Log current status
                logger.info(f"Strategy check completed at {datetime.now()}")
                
                # Wait before next iteration
                time.sleep(60)  # Check every minute

            except Exception as e:
                logger.error(f"Error in main loop: {e}")
                time.sleep(60)