# Concurrency Limiter
Implementing ideas from ["Stop Rate Limiting! Capacity Management Done Right" by Jon Moore](https://www.youtube.com/watch?v=m64SWl9bfvk&ab_channel=StrangeLoopConference).

<img width="1440" alt="Screenshot 2023-09-17 at 7 16 02 PM" src="https://github.com/PouriaAmini/concurrency-limiter/assets/64161548/e18f7fb8-81cf-4886-8111-5b38b0362404">

# Introduction
Concurrency Limiter is a proxy server that limits in-flight requests to 
servers and transfers back-pressure to clients during traffic spikes that exceed the server's capacity.
Unlike traditional rate limiters, which can be overly aggressive and reject 
requests that could otherwise be served, Concurrency Limiter adjusts its 
in-flight request (IFR)
limit using the [Additive increase/multiplicative decrease](https://en.wikipedia.org/wiki/Additive_increase/multiplicative_decrease) (AIMD) algorithm, increasing the transmission rate (window size), probing for usable bandwidth, until loss occurs, with packet loss serving as the signal. The multiplicative decrease is triggered when a timeout indicates
a packet loss.

# Project Overview
The project is based on a queueing Theory principal called 
[Little's Law](https://en.wikipedia.org/wiki/Little%27s_law), which states 
that the average number of items in a queue is equal to the average rate at 
which items arrive multiplied by the average time that an item spends in 
the queue, or mathematically: `L = λW`. In the context of this project, the
average number of items in the queue is the average number of in-flight 
requests (IFR), the average rate at which items arrive is the average number 
of requests per second (RPS), and the average time that an item spends in the
queue is the average response time (RT). Therefore, `IFR = RPS * RT`.

The server can only process a limited number of requests per
second, which is called the server's capacity. If the number of requests
exceeds the server's capacity, the server will start to queue the requests.
If the requests injected into the system faster than they are processed, the
queue will grow indefinitely, and the response time will increase. For example,
if the server has 7 workers and the average response time is 2 seconds, the
server's capacity is 7/2 = 3.5 RPS. So, the worker thread pool can pull 
through 3.5 requests per second, which is the rate at which requests are
drained from the queue. Now, if the client injects 5 requests per second into
the system, the queue will grow indefinitely, and the response time will
increase.

<img width="666" alt="Screenshot 2023-09-22 at 6 08 18 PM" src="https://github.com/PouriaAmini/concurrency-limiter/assets/64161548/4f213b69-969d-4878-ba28-61c06af2189b">

The goal of this project is to limit the number of in-flight requests to the
server to prevent the queue from growing indefinitely and to keep the response
time low.

- ### Proxy Server
  The proxy server will limit the number of in-flight requests to the
  server by limiting the number of requests that it forwards to the server. 
  It will also transfer back-pressure to the client by limiting the
  number of requests that it forwards to the client,
  monitor the response time of the server and adjust the number of in-flight
  requests to the server based on the response time. If the response time is
  low, the proxy server will increase the number of in-flight requests to the
  server. If the response time is high, the proxy server will decrease the
  number of in-flight requests to the server. The adjustment is done using the
    [Additive increase/multiplicative decrease](https://en.wikipedia.org/wiki/Additive_increase/multiplicative_decrease) (AIMD) algorithm.
  The proxy server is implemented as a Lua extension on an NGINX reverse proxy 
  instance.

  Here's a demo of how the proxy limits the client's request pressure and keeps
  the server safe from becoming overcrowded with requests and can serve as many
  requests as possible:


https://github.com/PouriaAmini/concurrency-limiter/assets/64161548/2827d88c-6521-4488-977b-fee4050c0ca1

    
- ### Client/Server
    The client and the server simulate a real-world production traffic 
  implemented in Go. The client can be configured as follows:
    - `--id`: The client's ID.
    - `--port`: The port that the client listens on.
    - `--targetPort`: The port that the client sends requests to.
    - `--rate`: The rate at which the client sends requests to the server.\
    The server can be configured with the following flags:
    - `--port`: The port that the server listens on.
    - `--rate`: The rate at which the server processes requests.
    - `--delay`: The delay that the server introduces to simulate a real-world
  
- ### Metrics
    The metrics from the client, server, and proxy server are scraped by
    Prometheus and visualized using Grafana. The Grafana dashboard can be
    accessed at `localhost:3000` with the username `admin` and the password
    `admin`.
