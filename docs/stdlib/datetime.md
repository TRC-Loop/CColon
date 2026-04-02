# datetime

The `datetime` module provides functions for working with dates, times, and timezones.

```
import datetime
```

## Functions

| Function | Description |
|---|---|
| `datetime.now()` | Returns current local time as a dict |
| `datetime.utcnow()` | Returns current UTC time as a dict |
| `datetime.timestamp()` | Returns current Unix timestamp (seconds) |
| `datetime.timestamp_ms()` | Returns current Unix timestamp (milliseconds) |
| `datetime.from_timestamp(int ts)` | Converts Unix timestamp to datetime dict |
| `datetime.format(dict dt, string fmt)` | Formats a datetime dict as a string |
| `datetime.parse(string s, string fmt)` | Parses a string into a datetime dict |
| `datetime.timezone(dict dt, string tz)` | Converts a datetime to another timezone |
| `datetime.diff(dict a, dict b)` | Returns difference in seconds (float) |
| `datetime.sleep(int ms)` | Sleeps for the given number of milliseconds |

## Datetime Dict

All datetime functions that return a time use a dict with these keys:

```
{
    "year": 2026,
    "month": 4,
    "day": 1,
    "hour": 14,
    "minute": 30,
    "second": 0,
    "weekday": "Wednesday",
    "timezone": "Local"
}
```

## Format Tokens

The `format` and `parse` functions use these tokens:

| Token | Meaning | Example |
|---|---|---|
| `%Y` | 4-digit year | `2026` |
| `%m` | 2-digit month | `04` |
| `%d` | 2-digit day | `01` |
| `%H` | 24-hour hour | `14` |
| `%M` | Minute | `30` |
| `%S` | Second | `00` |
| `%Z` | Timezone abbreviation | `UTC` |

## Examples

### Get current time

```
import console
import datetime

var dict now = datetime.now()
console.println("Year: " + now["year"].tostring())
console.println("Month: " + now["month"].tostring())
```

### Format and parse

```
import console
import datetime

var dict now = datetime.now()
var string formatted = datetime.format(now, "%Y-%m-%d %H:%M:%S")
console.println(formatted)

var dict parsed = datetime.parse("2026-01-15 10:30:00", "%Y-%m-%d %H:%M:%S")
console.println(parsed["year"].tostring())
```

### Timezones

```
import console
import datetime

var dict utc = datetime.utcnow()
var dict tokyo = datetime.timezone(utc, "Asia/Tokyo")
var dict ny = datetime.timezone(utc, "America/New_York")

console.println("UTC: " + datetime.format(utc))
console.println("Tokyo: " + datetime.format(tokyo))
console.println("New York: " + datetime.format(ny))
```

### Timestamps

```
import console
import datetime

var int ts = datetime.timestamp()
console.println("Unix timestamp: " + ts.tostring())

var dict dt = datetime.from_timestamp(ts)
console.println(datetime.format(dt))
```

### Time difference

```
import console
import datetime

var dict start = datetime.now()
datetime.sleep(1000)
var dict end = datetime.now()

var float diff = datetime.diff(end, start)
console.println("Elapsed: " + diff.tostring() + " seconds")
```
