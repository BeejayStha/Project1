# Project 1: Process Scheduler


## Description 
For this project we'll be building a simple process scheduler that takes in a file containing example processes, and outputs a schedule based on the three different schedule types:

- First-Come First-Serve (FC-FS)
- Shortest-Job-First (SJF)
- SJF-Priority
- Round-robin (RR)
- Assume that all processes are CPU bound (they do not block for I/O).

### Compile Process:
In order to compile I used Visual Studio Code. Then run command 
- go run main.go example_processes.csv

#### Sample Output
----------------------------------------------
            First-come, first-serve
----------------------------------------------
Gantt schedule
|   1   |   2   |   3   |
0       5       14      20

Schedule table
+----+----------+-------+---------+---------+------------+------------+
| ID | PRIORITY | BURST | ARRIVAL |  WAIT   | TURNAROUND |    EXIT    |
+----+----------+-------+---------+---------+------------+------------+
|  1 |        2 |     5 |       0 |       0 |          5 |          5 |
|  2 |        1 |     9 |       3 |       2 |         11 |         14 |
|  3 |        3 |     6 |       6 |       8 |         14 |         20 |
+----+----------+-------+---------+---------+------------+------------+
|                                   AVERAGE |  AVERAGE   | THROUGHPUT |
|                                    3.33   |   10.00    |   0.15/T   |
+----+----------+-------+---------+---------+------------+------------+
------------------------------------
          Shortest-job-first
------------------------------------
Gantt schedule
|   1   |   1   |   2   |   2   |   3   |   3   |
0       1       3       4       6       7       8

Schedule table
+----+----------+-------+---------+---------+------------+------------+
| ID | PRIORITY | BURST | ARRIVAL |  WAIT   | TURNAROUND |    EXIT    |
+----+----------+-------+---------+---------+------------+------------+
+----+----------+-------+---------+---------+------------+------------+
|                                   AVERAGE |  AVERAGE   | THROUGHPUT |
|                                    0.00   |    2.67    |   0.38/T   |
+----+----------+-------+---------+---------+------------+------------+
--------------------------------------
          SJFPrioritySchedule
--------------------------------------
Gantt schedule
|   1   |   1   |   3   |   2   |   2   |   3   |
0       5       10      16      25      34      40

Schedule table
+----+----------+-------+---------+---------+------------+------------+
| ID | PRIORITY | BURST | ARRIVAL |  WAIT   | TURNAROUND |    EXIT    |
+----+----------+-------+---------+---------+------------+------------+
|  1 |        2 |     5 |       0 |       5 |         10 |         10 |
|  2 |        1 |     9 |       3 |      22 |         31 |         34 |
|  3 |        3 |     6 |       6 |      28 |         34 |         40 |
+----+----------+-------+---------+---------+------------+------------+
|                                   AVERAGE |  AVERAGE   | THROUGHPUT |
|                                    24.00  |   37.33    |   0.07/T   |
+----+----------+-------+---------+---------+------------+------------+
----------------------
      Round-robin
----------------------
Gantt schedule
|   1   |   2   |   3   |   1   |   2   |   3   |   1   |   2   |   3   |   2   |   2   |
0       3       6       6       8       10      12      13      15      17      19      20

Schedule table
+----+----------+-------+---------+---------+------------+------------+
| ID | PRIORITY | BURST | ARRIVAL |  WAIT   | TURNAROUND |    EXIT    |
+----+----------+-------+---------+---------+------------+------------+
|  1 |        2 |     1 |       0 |      12 |         13 |         13 |
|  3 |        3 |     2 |       6 |       9 |         11 |         17 |
|  2 |        1 |     1 |       3 |      16 |         17 |         20 |
+----+----------+-------+---------+---------+------------+------------+
|                                   AVERAGE |  AVERAGE   | THROUGHPUT |
|                                    12.33  |   13.67    |   0.15/T   |
+----+----------+-------+---------+---------+------------+------------+
