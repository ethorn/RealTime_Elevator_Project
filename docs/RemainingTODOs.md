
Remaining TODOs
===========
Divided into more concrete tasks as we go

Remaining bugs/edge cases
----------
[] Flytte den første ifen i for selecten til inni select case peer update hvis mulig, unngår på den måten "globale" variabler
[] Flagging eller annen mekanisme som detekterer timeout og utløser reset
[] Funker obstruction som den skal? Har ikke testet den skikkelig. Vi trenger kanskje opprette en EB_Stuck?
[] Obstruction og andre hindringer -> oppdatere unservicable peers
[] En begrensning ved systemet vårt er at må endre en del kode ved å øke n elevators
[] lys som blir slått på når de ikke burde (tror jeg vet årsaken, må finne riktig sted å wipe ordre) 
[] Done? - Heisene respekterer ikke det at døra er åpen, prioriterer hall-calls midt i en cab call 
[] Done? - Stop og obstruct er jeg litt usikker på om gjør det de skal? Må også finne en måte å ta heisene unna nye ordre og eventuelt distribuere ordre på nytt om de står for lenge
[] Done? - Døren er asynch med bevegelser? Timeren er dobbelt så lang som den skal være? Eller på 4 s?
[] Done? - Ved init sendes ikke StateMessages, går vel fint? -> Sendes ved new peers nå?
[] PACKAGE LOSS testing og error tests for øvrig
[] Nevn med de andre at det å trykke på stoppknappen utlyser true deretter false
      "klasbo (Anders): 
      Bug-feature i simulatoren: Når en knapp trykkes, så a) sendes knappetrykket, og b) legges det til en timer event som trigger etter 200ms for å løfte knappen. Hvis man da trykker knappen igjen før 200ms (ved at man holder ned knappen og har keyboard hold-to-repeat), så vil ikke timer-eventen resettes, bare en ny en legges til.
      Bruk uppercase (Shift + char) for å holde nede knapper."
[] Fill in README.md as we go, introduce the different files
      Remember that other people will read your code. A readme file can be a good way to guide readers (including future self) to the relevant parts. You can still assume that the readers will know conventions for the programming language you're using. 
[] Rask overgang mellom n antall heiser og m antall etasjer, tror config-en er løsninga
      "Start with `1 <= n <= 3` elevators, and `m == 4` floors. Try to avoid hard-coding these values: You should be able to add a fourth elevator with no extra configuration, or change the number of floors with minimal configuration. You do, however, not need to test for `n > 3` and `m != 4`."
      "Note: You will be allowed to reinitialize your elevator system between testing with one, two and three elevators"
[] Masterheisen sender jevnlige meldinger om at den er masterheisen dersom ingenting skjer, er dette noe vi trenger å være obs på?

### Code review runthrough
The evaluation criteria below are an attempt to make evaluation criteria as precise as this kind of evaluation allows, but not perfect (or anywhere close). If you feel like the "technically correct" interpretation is against the intention of thecode review, follow your gut feeling.If your gut isn't helping you, try thinking about"minimizingaccidental complexity" or "maximizingmaintainability"
[]
[]
[]
[]
[]
[]
[]
[]
[]


### Hvis tid
[] Fiks edge caset med at elevatoren starter under 1. etg. eller over 4. etg., 
[] Sett opp en ekte config