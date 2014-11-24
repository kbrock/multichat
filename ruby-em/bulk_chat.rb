require 'em-websocket'

class Message
end

$debug_counter = 0
class Connection
  def initialize(ws)
    @name = ($debug_counter += 1)
    @ws = ws
    #@msg = Message.new
  end

  # one that actually writes
  def buffer msg
  end

  # store this in a buffer
  def send msg
    #@msg.merge(msg)
    @ws.send msg
  end

  def sync
    @ws.send @msg unless @msg.empty?
  end
end

class Hub
  attr_accessor :connections

  def initialize
    @connections = {}
    @counter = 0
  end

  def broadcast(msg)
    connections.each { |ws, c| c.send msg }
  end

  def register(ws)
    connections[ws] = Connection.new(ws)
  end

  # returns client (hash)
  def unregister(ws)
    connections.delete(ws)
  end

  def next_id
    @counter+=1
  end
end

hub = Hub.new

puts "listening on 8080"

# ServeWs, readPump
#EventMachine::WebSocket.start(:host => "0.0.0.0", :port => 8080) do |ws|
EM.epoll
trap("TERM") { EM.stop }
trap("INT")  { EM.stop }

EM.run do
  #EventMachine::WebSocket.run(:host => "0.0.0.0", :port => 8080) do |ws|
  EM.start_server("0.0.0.0", 8080, EventMachine::WebSocket::Connection, {}) do |ws|
    ws.onopen do
      hub.register(ws)
    end

    ws.onclose do
      hub.unregister(ws)
    end

    # actually in connection readPump
    ws.onmessage do |msg|
      hub.broadcast(msg)
    end
  end
end
