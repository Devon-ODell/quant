from hummingbot.client.config.global_config_map import global_config_map
from hummingbot.client.settings import AllConnectorSettings
from hummingbot.core.utils.async_utils import safe_ensure_future
from hummingbot.strategy.market_trading_pair_tuple import MarketTradingPairTuple
from hummingbot.strategy.avellaneda_market_making import (
    AvellanedaMarketMakingStrategy,
    AvellanedaMarketMakingConfigMap,
)

# Initialize the Hummingbot application
hb = HummingbotApplication()

# Setup configuration: replace with your actual API key and secret
api_key = 'zQgt66m+xwxD9Gh7f7TQiSC9smkcWsru2m3qx9YjzBWk9N9tYyl/NXeg'
api_secret = 'DfK0GNphJmtVi6EQtYcTge7j6OL4df6gT4BAcpTVN+5IY0/ErnjEDG2gJi2zadKmbT4hyipzmuCkOWDvr7axUw=='

exchange = 'kraken'
trading_pair = 'XRP-USD'

# Configure the exchange and API keys
config_map = global_config_map.get(exchange)
config_map.api_key = api_key
config_map.api_secret = api_secret
exchange_settings = AllConnectorSettings.get_exchange_settings(exchange)
connector = exchange_settings.connector_class(
    kraken_api_key=api_key,
    kraken_secret_key=api_secret,
)

# Setup the market trading pairs
market_info = MarketTradingPairTuple(exchange=connector, trading_pair=trading_pair, base_asset='XRP', quote_asset='USD')

# Strategy configuration
strategy = AvellanedaMarketMakingStrategy()
strategy_config_map = AvellanedaMarketMakingConfigMap.get_strategy_config_map()

# Customize your strategy settings below
strategy_config_map.min_spread = 0.01  # 1% minimum spread
strategy_config_map.inventory_risk_aversion = 1.0  # Inventory risk aversion
strategy_config_map.risk_factor = 0.5  # Adjust based on your risk preference
strategy_config_map.order_refresh_time = 30  # Order refresh time in seconds

# Initialize and start the strategy
strategy.init_params(
    config_map=strategy_config_map,
    market_info=market_info
)

# Start the strategy
safe_ensure_future(hb.start(strategy))

