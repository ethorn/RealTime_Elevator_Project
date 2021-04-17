Elevator Project
================

Summary (må reformuleres, men fikser det)
-------
Create software for controlling `n` elevators working in parallel across `m` floors.


Brief walkthrough of our setup
-------
We've gone with a master-slave network topology. Tne elevator always initially starts as a slave, but if no messages are received from the master within 6 seconds, one the connected slaves take over as the master. As evident by the single for-select loop inside the main function, most of the processes are the same regardless of hierachical status, but for the processes where the master-slave status does matter, the work is divided into two.