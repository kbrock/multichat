This is an experiment in a golang presentation chat server

Inspiration from:

- http://pomelo.netease.com/api.html
- http://www.slideshare.net/YorkTsai/jsdc2013-28389880
- https://www.youtube.com/watch?v=ysAZ_oqPOo0
- https://www.youtube.com/watch?v=Prkyd5n0P7k (group packets, don't send all packets)
- http://buildnewgames.com/optimizing-websockets-bandwidth/

Would like to send state, but instead of sending out a packet per packet in, would send fixed time:

- compress messages at transport layer (remove json overhead)
- send messages at small interval
- dedup packets so 
  a=1 b=2 a=3 ==> b=2 a=3

