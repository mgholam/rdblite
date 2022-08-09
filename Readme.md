# RDBLite
RaptorDB Lite is a simplified poor mans database engine based on my c# `RaptorDB Document Store Database`.

## How to use





## results

- using `gob` is 5x faster than `json` especially on slower cpus
- case insensitive search is 2x-3x slower
- filter with for loop -> not bad ~50ms worst case @ 100,000 items
- `TableInterface.GetID()` 25x faster than `reflect` for find by ID

### perf test 100,000 invoices

```sh
# powersave mode
2022/07/31 14:07:49 inv count 100000
2022/07/31 14:07:49 read invoices.json 600.839141ms
2022/07/31 14:07:49 inv count 100000
2022/07/31 14:07:49 read invoices.gob 107.685113ms       -> 5x faster
2022/07/31 14:07:49 start
2022/07/31 14:07:53 end: 4.710100761s
2022/07/31 14:07:53 start table1
2022/07/31 14:07:59 end process: 5.51559438s
2022/07/31 14:08:00 invoices.json written 514.967064ms
2022/07/31 14:08:00 invoices.gob written 128.920803ms    -> 4x faster
2022/07/31 14:08:00 filter count 1160
2022/07/31 14:08:00 invoices filter 15.034222ms
2022/07/31 14:08:00 filter invariant count 1160
2022/07/31 14:08:00 invoices filter 47.27317ms

# performance mode
2022/07/31 14:10:40 inv count 100000
2022/07/31 14:10:40 read invoices.json 189.062929ms
2022/07/31 14:10:40 inv count 100000
2022/07/31 14:10:40 read invoices.gob 36.745556ms        -> 5x faster
2022/07/31 14:10:40 start
2022/07/31 14:10:42 end: 1.649191125s
2022/07/31 14:10:42 start table1
2022/07/31 14:10:43 end process: 1.634122541s
2022/07/31 14:10:44 invoices.json written 184.957632ms
2022/07/31 14:10:44 invoices.gob written 42.808125ms     -> 4.3x faster
2022/07/31 14:10:44 filter count 1202
2022/07/31 14:10:44 invoices filter 11.618252ms
2022/07/31 14:10:44 filter invariant count 1202
2022/07/31 14:10:44 invoices filter 17.737819ms
```

