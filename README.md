Elevator Project
================

Summary (må reformuleres, men fikser det)
-------
Create software for controlling `n` elevators working in parallel across `m` floors.


Brief walkthrough of our setup
-------
We've gone with a master-slave network topology. Tne elevator always initially starts as a slave, but if no messages are received from the master within 6 seconds, one the connected slaves take over as the master. As evident by the single for-select loop inside the main function, most of the processes are the same regardless of hierachical status, but for the processes where the master-slave status does matter, the work is divided into two.

Single elevator mode is the natural way of operating. The difference becomes when a master sends new orders and when new orders are sent to others.
		How should elevators work while they have their own requests list?
		* internal chooseDirection algorithm
		* new cab call is added to internal requests state (and handled by internal chooseDir algo)
		* should serve cab calls when elevator is disconnected
		* This new cab call gets sent to others through updated state
		* new hall call is sent to master (with acks)
		if the master got the hall request, he checks if he is connected
		* Master then takes the hall request -> blackbox -> generates new requests for everyone. -> Send them
		Master always have the latest states of everyone, because they always send their state
		* should not serve hall requests when elev is disconnected
		* as the internal elevator gets a new state, it sends the new state to the master,
		which sends back an updated requests list (which includes cab requests)
		How to complete requests?