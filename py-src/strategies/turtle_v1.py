from jesse.strategies import Strategy
import jesse.indicators as ta
from jesse import utils


class TrendFollowingAI(Strategy):
    @property
    def ma50_big(self):
        return ta.sma(self.candles_6h, 50)

    @property
    def ma200_big(self):
        return ta.sma(self.candles_6h, 200)

    @property
    def ma20(self):
        return ta.sma(self.candles, 20)

    @property
    def ma50(self):
        return ta.sma(self.candles, 50)

    @property
    def macd(self):
        return ta.macd(self.candles, 12, 26, 9)

    @property
    def adx(self):
        return ta.adx(self.candles, 14)

    @property
    def atr(self):
        return ta.atr(self.candles, 14)

    def should_long(self) -> bool:
        # Big Trend Condition: Bullish
        if self.ma50_big > self.ma200_big and self.ma20 > self.ma50:
            # Entry Signals: MA20 crosses above MA50, MACD histogram > 0, ADX > 40
            if self.ma20 > self.ma50 and self.macd.hist > 0 and self.adx > 40:
                return True
        return False

    def go_long(self):
        entry_price = self.price
        qty = utils.size_to_qty(self.balance, entry_price)
        self.buy = qty, entry_price

    def should_short(self) -> bool:
        # Big Trend Condition: Bearish
        if self.ma50_big < self.ma200_big and self.ma20 < self.ma50:
            # Entry Signals: MA20 crosses below MA50, MACD histogram < 0, ADX > 40
            if self.ma20 < self.ma50 and self.macd.hist < 0 and self.adx > 40:
                return True
        return False

    def go_short(self) -> None:
        entry_price = self.price
        qty = utils.size_to_qty(self.balance, entry_price)
        self.sell = qty, entry_price

    def on_open_position(self, order):
        # Setting Stop Loss and Take Profit using ATR
        stop_loss = self.price - (self.atr * 1)
        take_profit = self.price + (self.atr * 2)
        self.stop_loss = self.position.qty, stop_loss
        self.take_profit = self.position.qty, take_profit

    def update_position(self):
        # Check exit conditions
        if self.ma20 < self.ma50 or self.macd.hist < 0:
            self.liquidate()

    def should_cancel_entry(self) -> bool:
        return False

    def filters(self) -> list:
        return []

    @property
    def candles_6h(self):
        return self.get_candles(self.exchange, self.symbol, '6h')
