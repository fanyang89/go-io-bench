# I/O benchmark

## Example

```bash
-> % go run io_bench.go generate -n samsung-9100pro -o fio-samsung-9100pro.sh --file-name /mnt/disk/test-file
# then run ./fio-samsung-990pro.sh 
```

```bash
-> % go run io_bench.go read-csv samsung-9100pro -t
Name                                Read IOPS  Read bw  Read lat     Write IOPS  Write bw   Write Lat    
samsung-9100pro_128k_randread_1     12.327 k   1.6 GB   80.944 us    0           -          -
samsung-9100pro_128k_randread_128   78.501 k   10 GB    1630.316 us  0           -          -            
samsung-9100pro_128k_randread_32    78.485 k   10 GB    407.544 us   0           -          -            
samsung-9100pro_128k_randwrite_1    0          -        -            31.097 k    4.0 GB     31.979 us    
samsung-9100pro_128k_randwrite_128  0          -        -            101.574 k   13 GB      1259.942 us  
samsung-9100pro_128k_randwrite_32   0          -        -            101.06 k    13 GB      316.459 us   
samsung-9100pro_128k_read_1         55.354 k   7.1 GB   17.917 us    0           -          -            
samsung-9100pro_128k_read_128       112.356 k  14 GB    1139.042 us  0           -          -            
samsung-9100pro_128k_read_32        112.374 k  14 GB    284.610 us   0           -          -            
samsung-9100pro_128k_write_1        0          -        -            30.958 k    4.0 GB     32.151 us    
samsung-9100pro_128k_write_128      0          -        -            102.089 k   13 GB      1253.622 us  
samsung-9100pro_128k_write_32       0          -        -            101.584 k   13 GB      314.850 us   
samsung-9100pro_16k_randread_1      19.801 k   317 MB   50.330 us    0           -          -            
samsung-9100pro_16k_randread_128    470.477 k  7.5 GB   271.888 us   0           -          -            
samsung-9100pro_16k_randread_32     388.787 k  6.2 GB   82.138 us    0           -          -            
samsung-9100pro_16k_randwrite_1     0          -        -            41.858 k    670 MB     23.712 us    
samsung-9100pro_16k_randwrite_128   0          -        -            403.437 k   6.5 GB     316.998 us   
samsung-9100pro_16k_randwrite_32    0          -        -            394.35 k    6.3 GB     80.861 us    
samsung-9100pro_16k_read_1          84.207 k   1.3 GB   11.727 us    0           -          -            
samsung-9100pro_16k_read_128        572.118 k  9.2 GB   223.583 us   0           -          -            
samsung-9100pro_16k_read_32         572.304 k  9.2 GB   55.771 us    0           -          -            
samsung-9100pro_16k_write_1         0          -        -            41.679 k    667 MB     23.841 us    
samsung-9100pro_16k_write_128       0          -        -            401.264 k   6.4 GB     318.769 us   
samsung-9100pro_16k_write_32        0          -        -            399.355 k   6.4 GB     79.909 us    
samsung-9100pro_256k_randread_1     9.573 k    2.5 GB   104.276 us   0           -          -            
samsung-9100pro_256k_randread_128   39.682 k   10 GB    3225.246 us  0           -          -            
samsung-9100pro_256k_randread_32    39.68 k    10 GB    806.254 us   0           -          -            
samsung-9100pro_256k_randwrite_1    0          -        -            22.421 k    5.7 GB     44.419 us    
samsung-9100pro_256k_randwrite_128  0          -        -            50.388 k    13 GB      2539.995 us  
samsung-9100pro_256k_randwrite_32   0          -        -            51.113 k    13 GB      625.870 us   
samsung-9100pro_256k_read_1         35.782 k   9.2 GB   27.799 us    0           -          -            
samsung-9100pro_256k_read_128       56.166 k   14 GB    2278.658 us  0           -          -            
samsung-9100pro_256k_read_32        56.123 k   14 GB    570.012 us   0           -          -            
samsung-9100pro_256k_write_1        0          -        -            22.381 k    5.7 GB     44.525 us    
samsung-9100pro_256k_write_128      0          -        -            50.561 k    13 GB      2531.367 us  
samsung-9100pro_256k_write_32       0          -        -            51.209 k    13 GB      624.719 us   
samsung-9100pro_4k_randread_1       26.134 k   104 MB   38.094 us    0           -          -            
samsung-9100pro_4k_randread_128     709.028 k  2.8 GB   180.333 us   0           -          -            
samsung-9100pro_4k_randread_32      690.793 k  2.8 GB   46.133 us    0           -          -            
samsung-9100pro_4k_randwrite_1      0          -        -            70.996 k    284 MB     13.910 us    
samsung-9100pro_4k_randwrite_128    0          -        -            491.073 k   2.0 GB     260.361 us   
samsung-9100pro_4k_randwrite_32     0          -        -            491.024 k   2.0 GB     64.881 us    
samsung-9100pro_4k_read_1           85.548 k   342 MB   11.541 us    0           -          -            
samsung-9100pro_4k_read_128         699.105 k  2.8 GB   182.942 us   0           -          -            
samsung-9100pro_4k_read_32          697.964 k  2.8 GB   45.701 us    0           -          -            
samsung-9100pro_4k_write_1          0          -        -            70.849 k    283 MB     13.965 us    
samsung-9100pro_4k_write_128        0          -        -            500.32 k    2.0 GB     255.613 us   
samsung-9100pro_4k_write_32         0          -        -            501.803 k   2.0 GB     63.556 us    
samsung-9100pro_8k_randread_1       21.01 k    168 MB   47.426 us    0           -          -            
samsung-9100pro_8k_randread_128     506.527 k  4.1 GB   252.523 us   0           -          -            
samsung-9100pro_8k_randread_32      424.707 k  3.4 GB   75.176 us    0           -          -            
samsung-9100pro_8k_randwrite_1      0          -        -            69.007 k    552 MB     14.313 us    
samsung-9100pro_8k_randwrite_128    0          -        -            460.007 k   3.7 GB     277.983 us   
samsung-9100pro_8k_randwrite_32     0          -        -            452.436 k   3.6 GB     70.460 us    
samsung-9100pro_8k_read_1           84.961 k   680 MB   11.623 us    0           -          -            
samsung-9100pro_8k_read_128         636.079 k  5.1 GB   201.056 us   0           -          -            
samsung-9100pro_8k_read_32          638.208 k  5.1 GB   49.970 us    0           -          -            
samsung-9100pro_8k_write_1          0          -        -            68.974 k    552 MB     14.347 us    
samsung-9100pro_8k_write_128        0          -        -            468.789 k   3.8 GB     272.822 us   
samsung-9100pro_8k_write_32         0          -        -            466.603 k   3.7 GB     68.366 us
```
