Elevator Project
================

Summary 
-----------

Ojective: Create software for controlling `n` elevators working in parallel across `m` floors.
Aim: Create a distributed fault-tolerant elevator system


Brief walkthrough of the main setup
-----------
For specific implementation details we refer the reader to (hopefully) extensive and informative comments associated with the code.

We have a master-slave network topology. An elevator always initially starts as a slave, but if no messages are received from the master within 6 seconds, one the connected slaves take over as the master. A slave can become master without losing any orders. As evident by the single for-select loop inside the main function, most of the processes are the same regardless of hierachical status, i.e. single elevator mode is the natural way of operating. The difference becomes when a master sends new orders and when new orders are sent to others.

In single elevator mode, each elevator has their own requests list. Any new cab call in the elevator is added to the internal request list and handled by an internal Choose Direction algorithm. The way single elevator mode is implemented, it should serve cab calls even when the elevator is disconnected.Information on new cab calls is sent to other elevators through an updated elevator state message, which naturally contains the updated internal request list. In single elevator mode, an elevator is free to reject hall orders (quoting Anders Pettersen's confirmation on Discord's general channel April 19, 7:28 PM)

New hall calls are sent to the master, and it's the master that distributes all hall orders. If the master receives a hall request, it checks if it is connected. Master then takes the hall request and sends it to the order logic module, which generates a new request for an elevator, thereafter sending it to the relevant elevator (TODO teknisk sett til alle?). Order logic is the module that connects all the order logic for the different elevators together, including clearing requests, designating orders and redistributing orders.

The master always has the latest states of every elevator, because they always send their state to the master. Hall calls are only distributed to connected elevators, and when an elevator is disconnected all of the disconnected elevator's hall orders are redistributed among the still connected peers through the order logic module. 

Since cab calls are handled internally for each elevator, they are stored locally on a file. Any uncompleted cab orders for an elevator elevator upon (re)initialisation are retrieved upon initialisation from a back up file of the cab orders. This file constantly contains a an updated list of remaining cab orders. Hall orders automatically do not become saved to the finalised file because as mentioned in the previous paragraph the hall orders are instantly redistributed upon the remaining connected elevators. The list of uncompleted cab orders inside the file is constantly up to date since every time a new state message is received, this list is written to the back up file. 

TODO: dobbeltsjekk dette:
In the case of obstruction in an elevator, the elevator is immediately registered as a "lost", i.e. disconnected elevator) and its hall orders are instantly redistributed, and the elevator program is registered as stuck. It may receive hall orders, but these are immediately redistributed to the other elevators that are actually. connected. If the elevator remains stuck over a longer time period, the elevator is forced to reset through the process pair method we've implemented.

The stop button is completely neglected, per specification allowances.


The different packages
-----------
Elevator: 
Basically a single struct containing information for an indiviual elevato, the code should speak for itself hopefully. TODO: nevne dette?: inherits states from ElevatorStates | passes states to ElevatorStates

Single elevator requests:
Functionality for single elevator mode.

Finite state machine (fsm):
Each elevator is a state machine, all state changes happen inside this package. Also includes timers.

Order logic:
Cab calls are handled locally, while an order logic algorithm handles hall calls. The order logic algorithm assigns only the new request, and it does so using a cost function based on _time until completion/idle_, where each individual elevator's total queue. In other words, for each available elevator the hypothetical cost of the new unassigned hall request is calculated by adding it to the existing workload (order queue, which takes into considering both cab and hall calls) of the elevator and simulating the execution. For this simulation we use the functions that we already have from the single elevator algorithm. A hall request is assigned to a particular elevator for the duration of the lifetime of that request, given that the elevator does not disconnect from the network or get stuck. 

Elevator I/O (elevio):
The elevator driver, handles input/output for buttons, lights, the door and the elevator motor. Contains no new functionality package from the provided one.

Network: 
Nothing has been adjusted in this package from the provided one. Communication between goroutines is through channels, while communication between elevators is through UDP broadcasting. When passing messages, acknowledgements are used to handle lost, corrupt, or multiple messages between elevators (TODO: blir vel feil å si multiple?)

Communication:
Contains functions for sending information through , for instance state updates, hall requests and cleared order. all acknowledgement implementation is implemented in this package, besides a acknowledge message handler in the fsm package.

Cab backup: 
Ensures reading from and writing to a file all incompleted cab orders for an elevator.

Config:
The package name speaks for itself, not as expansive as it could be but that would require extensive tinkering with code with the risk of causing more harm than good.


Underlying assumptions
-----------
We assume the user will be reasonable when it comes to using the elevator system, for instance with regards to not spamming the elevator buttons, pressing the buttons with a reasonable time space between them, exiting the elevator in a normal manner, and so on.
	- Therefore if somebody/something is obstructing an elevator, that elevator is immediately "unavailable"/lost to the rest of the elevators and all of its hall requests are therefore instantly redistributed. This still falls within the specification that "The obstruction switch should substitute the door obstruction sensor inside the elevator", but it is not overkill to simply shut that elevator out of the system.
	Instant redistribution on obstruction is within specification.

We assume the obstruction switch will not be flipped on all three elevators at the same time.