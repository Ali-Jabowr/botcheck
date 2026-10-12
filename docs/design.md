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

verify takes a resolver interface, so tests use a fake DNS instead of the real one.

## Alternatives considered

1. **Trusting the user-agent header alone.** Rejected: a scraper can pretend to be GPTBot, so the count for GPTBot will include fake counts.
2. **Reverse DNS for every request.** Rejected: each lookup is a network round trip cost roughly tens of milliseconds, so a 100,000-line log would take about 50 minutes, IP-ranges avoid this by checking the in memory cached IP ranges.
3. **A third-party bot-detection API.** Rejected: botcheck would send the request to that third-party service, so when the API servers of that service are off, botcheck will stop recording whether it's real bot, spoofed or doesn't claim to be a known bot.


## Risks / open questions

1. Many lines need DNS. Some bots don't publish ranges. I will reverse DNS lookup and cache the result per IP, because a big log has the same bot IPs on thousands of lines.
2. **DNS is slow or hangs.** A lookup can take a long time, I will set a timer for each lookup, and when it runs out the answer is couldn't check (timeout or error). Open: how long the timeout is (slice 4).

3. **Published ranges go stale.** Bots add new IPs, and old ranges would mark real bots as spoofed (mistake B). I refetch the ranges daily and keep the last-known-good list if a fetch fails or looks wrong (an error, an empty list, or much smaller than before). The age of the ranges is shown in the report and on `/health`, and becomes a loud problem after 7 days.
4. **If short on time:** concurrency (slice 5) is cut first, because the report still works sequentially thanks to the per-IP cache. Milestone 3 then moves to rung 4. **At 10× load** (1M lines, 10× the distinct IPs), DNS round trips are the first thing to break, and concurrency stops being optional.

## How we'll know it works

**The number: correctness on a known-answer log.** A real log can't grade botcheck, because nobody knows which lines are truly fake. So I generate a fake 100k-line nginx log where I already know the right answer for every line:
- real bot lines (bot user-agent, IP taken from the bot's published ranges)
- planted spoofed lines (bot user-agent, random IP)
- planted timeout lines (IPs whose lookup I force to time out)
- browser lines (Chrome user-agent)

**Target:**
1. 100% of planted fakes are flagged (spoofed or couldn't check), and 0 are marked real. Letting a fake through is the worst mistake, because a fake bot gets in for free.
2. 0 real bots are marked spoofed.
3. "Couldn't check" is allowed only on the lines I planted as timeouts. Otherwise a broken DNS setup, or a bug that returns "couldn't check" for everything, would pass the test.
