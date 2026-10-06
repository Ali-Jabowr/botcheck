## Context
so what is happening is that, bots and crawlers fetch pages on the internet all the time.
the user can't tell whether the request come from a real bot or from scraper, a scraper can pretend to be GPTBot and a bot can pretend to be browser, the ip check can't catch the second one.


## Goals
from an nginx access log:
1. a count of requests per bot, verified or spoofed.
    - GPTBot, ClaudeBot.

from one IP + User-agent over HTTP:
2. Whether it's a real bot, spoofed, couldn't check (timeout or error) and doesn't claim to be a known bot.


## Non-goals

1. Can't catch a bot pretending to be browser.
2. Doesn't block it that's tollgate.
3. Doesn't sit in front of the site; it reads logs after the fact.


## Design

```
 nginx log ──▶ [ read ] ──▶ [ parse ] ──┐
                                        ├──▶ [ identify ] ──▶ [ verify ] ──┬──▶ count
 ip + user-agent http ──────────────────┘          │                       └──▶ whether it's real bot, spoofed, couldn't check, doesn't claim to be a known bot
                                                   │
                                                   └── browser lines ──▶ log: dropped from the count
                                                                         http: doesn't claim to be a known bot
```

identify and verify are the shared core: the log path ends in a count, and the HTTP path ends in one answer.
Browser lines skip verify, the line will be written as Chrome user-agent so it doesn't claim to be a known bot.
verify checks the published IP ranges where they exist, otherwise reverse DNS with forward-confirm step, look up the hostname again and check it points back to the same IP, if it is timeout gives so "couldn't check" .

## Alternatives considered

1. **Trusting the user-agent header alone.** Rejected: a scraper can pretend to be GPTBot, so the count for GPTBot will include fake counts.
2. **Reverse DNS for every request.** Rejected: each lookup is a network round trip cost roughly tens of milliseconds, so a 100,000-line log would take about 50 minutes, IP-ranges avoid this by checking the in memory cached IP ranges.
3. **A third-party bot-detection API.** Rejected: botcheck would send the request to that third-party service, so when the API servers of that service are off, botcheck will stop recording whether it's real bot, spoofed or doesn't claim to be a known bot.
