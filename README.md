Elevator Project
================

Summary 
=========
Create software for controlling `n` elevators working in parallel across `m` floors.


Brief walkthrough of our setup
========

Single elevator mode is the natural way of operating. 

In single elevator mode, each elevator has their own requests list. 

We have a master-slave network topology. Tne elevator always initially starts as a slave, but if no messages are received from the master within 6 seconds, one the connected slaves take over as the master. As evident by the single for-select loop inside the main function, most of the processes are the same regardless of hierachical status. The difference becomes when a master sends new orders and when new orders are sent to others.


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


● Master polls updated information from slaves
● Master designates orders
● Slave can become Master without losing orders

New orders
● Acknowledged and stored
● Decentralized backup
Single elevator mode
● Cost function
● Only accepts cab-calls



The different packages
============
The network package: Nothing has been adjusted in this package, we're running with a UDP

Communication between goroutines is through channels, while communication between elevators is through UDP broadcasting. When passing messages, acknowledgements are used to handle lost, corrupt, or multiple messages between elevators (TODO: blir vel feil å si multiple?)


Order logic:
Cab calls are handled locally, while an order logic algorithm handles hall calls. The order logic algorithm assigns only the new request, and it does using a cost function based on time until completion, where each individual elevator's total queue. In other words, for each available elevator the hypothetical cost of the new unassigned hall request is calculated by adding it to the existing workload (order queue) of the elevator and simulating the execution. For this simulation we use the functions that we already have from the single elevator algorithm.

 is taken into consideration (in other words, both cab and hall calls are considered). A hall request is assigned to a particular elevator for the duration of the lifetime of that request, given that the elevator does not disconnect from the network or get stuck. 

From this we can calculate the cost of the new unassigned hall request by adding it to the existing workload and simulating the execution of the elevator. For this simulation we use the functions that we already have from the single elevator algorithm: Choose Direction, Should Stop, and Clear Requests At Current Floor:
Alternative 1.1: Time until completion/idle

Note that in order to reuse the function for clearing requests in a simulated context, we need to make sure it does not actually perform any side effects on its own. Otherwise, the simulation run might actually remove all the orders in the system, turn off lights, and so on.

The suggested modification is giving requests_clearAtCurrentFloor a second argument containing a function pointer to some side-effect, which lets us pass some function like "publish the clearing of this order", or in the case of our cost function - "do nothing", which is exactly what we want. (For most sensible modern languages, the passed-in function would be a lambda, or some other thing that lets you capture the enclosing scope, so the floor parameter to the inner function is probably not necessary)

If you are really scared of function pointers (or the designers of your language of choice were scared and didn't implement it), you can just make two functions (one simulated & one real), or pass a "simulate" boolean. Just make sure that the simulated behavior and real behavior are the same.



Disconnect/Stop
● Orders are rearranged on functioning elevators
● Master assumes no orders will be served
Reconnect
● Pulls updated info from new master


Underlying assumptions
===========
- We assume the user will be reasonable when it comes to using the elevator system, for instance with regards to calling the elevator, exiting the elevator in a normal manner, and so on.
	- Therefore if somebody/something is obstructing an elevator, that elevator is immediately "unavailable"/lost to the rest of the elevators and all of its hall requests are therefore instantly redistributed. This still falls within the specification that "The obstruction switch should substitute the door obstruction sensor inside the elevator", but it is not overkill to simply shut that elevator out of the system.
- We assume the obstruction switch will not be flipped on all three elevators at the same time.



"Instant redistribution on obstruction is within spec. (So is redistribution for any reason at any time, as long as nothing in Multiple elevators should be more efficient than one is violated)"