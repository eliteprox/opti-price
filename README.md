# Optiprice
Designed to be used with crontab at your preferred interval.
Example, every 15 minutes
`*/15 * * * * ~/optiprice/./optiprice`

Requires change to point to your prometheus database (this will be updated to read from `/status` endpoint instead soon)

```
Sample config.json
{
    "high_stream_count": 10,
    "target_stream_count": 8,
    "low_price": 100,
    "high_price": 600,
    "price_increment": 50
}
```

- Each interval that goes over the `high_stream_count` will incur a price increase by the `price_increment`
- Each interval that goes under the `target_stream_count` will incur a price decrease by the `price_increment`
- Each interval that is between `target_stream_count` and `high_stream_count` will incur no price change
- The price will not go below `low_price` or above `high_price`

