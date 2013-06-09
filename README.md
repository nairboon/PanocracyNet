
# AnEvoNet
AnEvoNet shall become an self evolving peer-to-peer platform, enabling the development of novel & revolutionary p2p applications.

## Core idea
It is inspired by the idea of [artificial social networks][1] by David Hales. ASNs achieve cooperation through playing the Prisoners Dilemma (PD) game, where peers copy and mutate the strategy of more successful peers.

In AnEvoNet the available strategies will not be cooperation or betrayal but a wide range of algorithms (topology, search, compression, etc.). The best performing will eventually be adopted by the whole p2p network, thus leading to an optimal performance. 

Additionally there won't be a single fitness function like the proposed number of files downloaded, but instead there'll be separate optimizations for each known use case i.e. file sharing can tolerance high latency but needs a high throughput while instant messaging requires the opposite.
By measuring the access patterns of individual applications the core system will be able to optimize the strategy per application.

## What is it good for?
I'm sure you can come up with some use cases just think about that eventually you'll have low-level APIs to

- store an infinite amount of content
- access the same content
- communicate in real time through low latency channels
- encryption?
- access to a distributed computing environment (think a small boinc)

and maybe high level APIs for

- recommender systems for your favourite content
- collaborative databases

## Implementation

### Intro to p2p systems
The use of a p2p system is based on the number of active peers and on the effectiveness of the implementation. In other words your p2p application is either slow because there are too few peers or because your implementation sucks. The number of peers on the other hand is based on the use of the p2p system and the ease to use it. 

So if your application is hard to use from an end user perspective, either because the UI is unintuitive or because installing the app is impossible for non techies, very few user will use the system thus the performance and the use will degrade no matter how good the implementation is.

But if your application is very easy to use i.e. through a website you might get many users but their user experience won't be as good as with a solid implementation because the webapp is very limited compared to a native application.

AnEvoNet tries to escape this dilemma through a combination of both ways. A lightweight web application can show the possibilities and attract users while a solid native implementation ensures performance and allows power users to get the most out of the network.

Wait! P2P platform and webapplication??? Yes exactly, P2P in the browser has become possible in the last months through technologies like WebSockets and WebRTS.

AnEvoNet will achieve this through defining an extendable protocol, a in-browser implementation and a native high-performance implementation, both being able to communicate.

### Plans
There are different components to develop, none exists as of right now.

#### Meta Protocol
The meta protocol should be versioned and extendable, so that each peer can choose his favorite strategy that is compatible with his peer, i.e. on native-browser connections.

#### Browser Implementation
The browser implementation will be written in Javascript or a superset. 
#### Native Implementation
For the native implementation I intend to use the concurrent go language.

#### Example Applications
...



[1]: http://arxiv.org/abs/cs/0509037


