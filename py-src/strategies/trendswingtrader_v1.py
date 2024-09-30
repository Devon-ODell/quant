from jesse.strategies import Strategy, cached
import jesse.indicators as ta
from jesse import utils


class TrendSwingTrader(Strategy):
    @property
    def adx(self):
        return ta.adx(self.candles) > 25

    @property
    def trend(self):
        e1 = ta.ema(self.candles, 21)
        e2 = ta.ema(self.candles, 50)
        e3 = ta.ema(self.candles, 100)
        if e3 < e2 < e1 < self.price:
            return 1
        elif e3 > e2 > e1 > self.price:
            return -1
        else:
            return 0

    def should_long(self) -> bool:
        return self.trend == 1 and self.adx

    def go_long(self):
        entry = self.price
        stop = entry - ta.atr(self.candles) * 2
        qty = utils.risk_to_qty(self.available_margin, 5, entry, stop, fee_rate=self.fee_rate) * 2
        self.buy = qty, entry

    def should_short(self) -> bool:
        return self.trend == -1 and self.adx

    def go_short(self):
        entry = self.price
        stop = entry + ta.atr(self.candles) * 2
        qty = utils.risk_to_qty(self.available_margin, 5, entry, stop, fee_rate=self.fee_rate) * 2
        self.sell = qty, entry

    def should_cancel_entry(self) -> bool:
        return True

    def on_open_position(self, order) -> None:
        if self.is_long:
            self.stop_loss = self.position.qty, self.price - ta.atr(self.candles) * 2
            self.take_profit = self.position.qty/2, self.price + ta.atr(self.candles) * 3
        elif self.is_short:
            self.stop_loss = self.position.qty, self.price + ta.atr(self.candles) * 2
            self.take_profit = self.position.qty/2, self.price - ta.atr(self.candles) * 3

    def on_reduced_position(self, order) -> None:
        if self.is_long:
            self.stop_loss = self.position.qty, self.position.entry_price
        elif self.is_short:
            self.stop_loss = self.position.qty, self.position.entry_price

    def update_position(self) -> None:
        if self.reduced_count == 1:
            if self.is_long:
                self.stop_loss = self.position.qty, max(self.price - ta.atr(self.candles) * 2, self.position.entry_price)
            elif self.is_short:
                self.stop_loss = self.position.qty, min(self.price + ta.atr(self.candles) * 2, self.position.entry_price)
