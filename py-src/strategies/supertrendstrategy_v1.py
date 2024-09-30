from jesse.strategies import Strategy
import jesse.indicators as ta
from jesse import utils


class davidSuperTrend(Strategy):
    def should_long(self) -> bool:
        return self.slow_super_trend < self.price and self.fast_super_trend < self.price and self.macd > 0 and self.vortex.plus > self.vortex.minus and self.stiffness

    def go_long(self):
        qty = utils.size_to_qty(self.available_margin, self.price, fee_rate=self.fee_rate)
        self.buy = qty, self.price

    def should_short(self) -> bool:
        return self.slow_super_trend > self.price and self.fast_super_trend > self.price and self.macd < 0 and self.vortex.plus < self.vortex.minus and self.stiffness

    def go_short(self):
        qty = utils.size_to_qty(self.available_margin, self.price, fee_rate=self.fee_rate)
        self.sell = qty, self.price

    def should_cancel_entry(self) -> bool:
        return True

    def on_open_position(self, order) -> None:
        atr_band = ta.atr(self.candles)
        if self.is_long:
            self.take_profit = self.position.qty, self.position.entry_price + 3 * atr_band
            self.stop_loss = self.position.qty, self.position.entry_price - 2 * atr_band
        elif self.is_short:
            self.take_profit = self.position.qty, self.position.entry_price - 3 * atr_band
            self.stop_loss = self.position.qty, self.position.entry_price + 2 * atr_band

    def watch_list(self) -> list:
        return [
            ('super_trend', self.super_trend),
            ('macd', self.macd),
            ('vortex', self.vortex)
        ]

    @property
    def fast_super_trend(self):
        return ta.supertrend(self.candles, period=27, factor=13).trend

    @property
    def slow_super_trend(self):
        return ta.supertrend(self.candles, period=27, factor=15).trend

    @property
    def macd(self):
        return ta.macd(self.candles, fast_period=14, slow_period=28, signal_period=11).macd

    @property
    def vortex(self):
        return ta.vi(self.candles, period=14)

    @property
    def stiffness(self):
        stiffness = ta.stiffness(self.candles, threshold=40)
        return stiffness.stiffness > stiffness.threshold
