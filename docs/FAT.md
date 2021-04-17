Concrete specifications/FAT suggestions
==============

[]
[]
[]
[]
[]


### No orders are lost
[] Once the light on a hall call button is turned on, an elevator should arrive at that floor
[] Similarly for a cab call, but only the elevator at that specific workspace should take the order
[] Handling network packet loss, losing network connection entirely, software that crashes, and losing power - both to the elevator motor and the machine that controls the elevator
[] For cab orders, handling loss of power/software crash implies that the orders are executed once service is restored
         - Network packet loss is not an error, and can occur at any time
[] The time used to detect these failures should be reasonable, ie. on the order of magnitude of seconds (not minutes)
[] If the elevator is disconnected from the network, it should still 
        [] Serve all the currently active orders (ie. whatever lights are showing)
        [] Keep taking new cab calls, so that people can exit the elevator even if it is disconnected from the network
[] The elevator software does not require reinitialization (manual restart) after intermittent network or motor power loss

### Multiple elevators should be more efficient than one
[] Orders are distributed across the elevators in a reasonable way
        [] Ex: If all three elevators are idle and two of them are at the bottom floor, then a new order at the top floor should be handled by the closest elevator (ie. neither of the two at the bottom).
        [] Other exps: 
[] It is clear that the elevators are cooperating and communicating.
 
### An individual elevator should behave sensibly and efficiently
[] The hall "call upward" and "call downward" buttons behave differently
        [] Ex: If the elevator is moving from floor 1 up to floor 4 and there is a downward order at floor 3, then the elevator should not stop on its way upward, but should return back to floor 3 on its way down
        [] Other exs:


### The lights and buttons should function as expected
[] The hall call buttons on all workspaces let you summon an elevator
[] Under normal circumstances, the lights on the hall buttons should show the same thing on all workspaces 
[] Under circumstances with high packet loss, at least one light must work as expected
[] The cab button lights are not shared between elevators
[] The cab and hall button lights turn on as soon as is reasonable after the button has been pressed
       - You are allowed to expect the user to press the button again if it does not light up
[] The cab and hall button lights turn off when the corresponding order has been serviced
[] The "door open" lamp is used as a substitute for an actual door, and as such is not switched on while the elevator is moving
[] The duration for keeping the door open is in the 1-5 second range
[] The obstruction switch substitutes the door obstruction sensor inside the elevator. The door does not close while it is obstructed.
 
Start with `1 <= n <= 3` elevators, and `m == 4` floors. Try to avoid hard-coding these values: You should be able to add a fourth elevator with no extra configuration, or change the number of floors with minimal configuration. You do, however, not need to test for `n > 3` and `m != 4`.

### Fault tolerance 
(overlapper litt med No orders are lost)
The elevator system should adher to specification when faults are introduced. 
[]
[]
[]
[]
[]  Your elevator system must adher to specification even with simulated packet loss on your network adapter (not on localhost). 
        To simulate packet loss we will use the following terminal command:
                sudo iptables -A INPUT -p tcp --dport 15657 -j ACCEPT
                sudo iptables -A INPUT -p tcp --sport 15657 -j ACCEPT
                sudo iptables -A INPUT -m statistic --mode random --probability 0.2 -j DROP
        To remove packet loss:
                sudo iptables -F


### Etc.
[] 


Unspecified behaviour
---------------------
Some things are left intentionally unspecified. Their implementation will not be tested, and are therefore up to you.

Which orders are cleared when stopping at a floor
 - You can clear only the orders in the direction of travel, or assume that everyone enters/exits the elevator when the door opens
 
How the elevator behaves when it cannot connect to the network (router) during initialization
 - You can either enter a "single-elevator" mode, or refuse to start
 
How the hall (call up, call down) buttons work when the elevator is disconnected from the network
 - You can optionally refuse to take these new orders
 
What the stop button does
   - The stop button functionality (if/when implemented) is up to you

   
Permitted assumptions
---------------------

The following assumptions will always be true during testing:
 1. At least one elevator is always working normally
 2. No multiple simultaneous errors: Only one error happens at a time, but the system must still return to a fully operational state after this error
    - Recall that network packet loss is *not* an error in this context, and must be considered regardless of any other (single) error that can occur
 3. No network partitioning: There will never be a situation where there are multiple sets of two or more elevators with no connection between them
 4. Cab call redundancy with a single elevator is not required
    - Given assumptions **1** and **2**, a system containing only one elevator is assumed to be unable to fail
   




Fra TilpDat-kriteriene
---------------------
Generelle kriterier:
	· Ved oppstart skal heisen alltid komme til en definert tilstand. En definert tilstand betyr at styresystemet vet hvilken etasje heisen står i.
	· Om heisen starter i en udefinert tilstand, skal heissystemet ignorere alle forsøk på å gjøre bestillinger, før systemet er kommet i en definert tilstand.
	· Heissystemet skal ikke ta i betraktning urealistiske startbetingelser, som at heisen er over fjerde etasje, eller under første etasje idet systemet skrus på.
	· Det skal ikke være mulig å komme i en situasjon hvor en bestilling ikke blir tatt. Alle bestillinger skal betjenes, selv om nye bestillinger opprettes.
	· Heisen skal ikke betjene bestillinger fra utenfor heisrommet om heisen er i bevegelse i motsatt retning av bestillingen.
	· (avhenger av CV config vi har, men hvis vi har CV All må denne her gjelde:)
	Når heisen først stopper i en etasje, skal det antas at alle som venter i etasjen går på eller av. Dermed skal alle ordre i etasjen være regnet som ekspedert.
	· Om heissystemet ikke har noen ubetjente bestillinger, skal heisen stå stille.
	· Heisen skal aldri kjøre utenfor området definert av første- og fjerde etasje.
	· Det skal ikke være nødvendig å periodisk starte programmet på nytt som følger av eksempelvis udefinert oppførsel, at programmet krasjer, eller minnelekkasje

Kriterier m.t.p. lysene:
	· Om en bestillingsknapp ikke har en tilhørende bestilling, skal lyset i knappen være slukket.
	· Når heisen er i en etasje skal korrekt etasjelys være tent.
	· Når heisen er i bevegelse mellom to etasjer, skal etasjelyset til etasjen heisen sist var i være tent.
	· Når en bestilling gjøres, skal lyset i bestillingsknappen lyse helt til bestillingen er utført. Dette gjelder både bestillinger inne i heisen, og bestillinger utenfor.

Kriterier m.t.p. døren:
	· Om heisen ikke har ubetjente bestillinger, skal heisdøren være lukket.
	· Når heisen ankommer en etasje det er gjort bestilling til, skal døren åpnes i 3 (edit: 1-5) sekunder, for deretter å lukkes.

Kriterier m.t.p. obstruksjonsbryteren:


Kriterier m.t.p. stoppknappen (opp til oss hvilke vi implementerer):
	· Hvis stoppknappen trykkes mens heisen er i en etasje, skal døren åpne seg. Døren skal forholde seg åpen så lenge stoppknappen er aktivert, og ytterligere 3 sekunder etter at stoppknappen er sluppet. Deretter skal døren lukke seg.
	· Om stoppknappen trykkes, skal heisen stoppe momentant.
	· Så lenge stoppknappen holdes inne, skal heisen ignorere alle forsøk på å gjøre bestillinger.
	· Etter at stoppknappen er blitt sluppet, skal heisen stå i ro til den får nye bestillinger.



Code review criteria
---------------------
"[...] try thinking about "minimizing accidental complexity" or "maximizing maintainability"."
"The bullet points for each question are lists of examples of things to look for, but are not exhaustive - the code does not have to satisfy all (or any) of the bullet points - instead, stay within the theme described in the title."

The main function:
- Components:
  [] The entry point documents what components/modules the system consists of. I.e., you can see what threads/classes are initialized.
- Dependencies:
  [] The entry point documents how these components are connected.
     [] You can see how different components interact and depend on each other. This would imply making channels, thread ID's, or object-pointers here, and explicitly passing them to the relevant components
     [] If there are any global variables, their use is immediately clear and are their names are truly excellent
- Functionality:
  [] The reader knows where to look in order to find out how the system is designed.
     For instance:
     [] If it is master-slave or peer-to-peer
     [] How any acknowledgment procedure works
     [] How any order assignment works
     [] How orders for this elevator are executed
     [] How orders are backed up

The individual modules from the "outside" (i.e., the header file, the public functions, the list of channels or types the process reads from, etc.):
- Coherence:
  [] The module appears to deal with only one subject. 
     E.g. if you find things that deal with orders/requests, you shouldn't also find things that deal with network
- Completeness:
  [] The module appears to deal with "everything" concerning that subject.
     [] There are no cases where the interfaces show an obvious lack of functionality
     [] E.g. if there is an "add order", there should probably also be a way to "remove order(s)"

The individual modules from the "inside" (i.e., the contents/bodies of the functions, select- or receive-statements, etc.):
- State:
  [] State is maintained in a structured and local way
     [] A thread, process or object captures all interactions with a piece of state
     [] Writable state is not passed around to other modules, only values or copies
- Functions:
  [] Routines are as pure as possible
     [] Routines do not modify variables outside their scope
     [] If there are any variables with a scope larger than the routine, it is trivial to find out what their scope is, and the variables are very easy to keep track of
- Understandability:
  [] Each body of code is easy to follow
     [] You can see what it does, and you can see that it is correct
     [] E.g. nesting levels are kept under control, local variables have names that don't confuse you, etc.

The interaction between modules (i.e., how information flows from one module to the next):
[] For example, try to trace an event like a button press, and follow the information from its source (something reading the elevator hardware) to its destination (some other elevator starts moving)
- Traceability:
  [] The flow of information can be traced easily
     [] A process or object that changes its state has a clear origin point for why it changed its state. 
        Think debugging scenarios like "why does this variable have this value now?"
- Direction:
  [] The information flows (mostly) in one direction, from one module to the next.
     [] E.g. if A calls into B, then B does not call back into A again - usually

Details (i.e. the contents/bodies of the functions, select- or receive-statements, etc.):
- Comments
  [] The comments the readers finds are useful
     [] The comments are not just a repetition of the code
     [] "If there are no comments and you feel no comments were necessary, award the point" - direkte sitering fra code quality evaluation criteria
- Naming
  [] The names of modules, functions, etc. help the reader navigate the code
     [] The reader is never mislead by a vague or incorrect name