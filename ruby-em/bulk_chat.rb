require 'em-websocket'

$counter = 1

class Message
end

class Connection
end

class Hub
  attr_accessor :connections

  def initialize
    @connections = []
    @counter = 0
  end

  def broadcast(msg)
    connections.each { |s| s[:socket].send msg }
  end

  def register(ws)
    connections.push(id: next_id, socket: ws)
  end

  # returns client (hash)
  def unregister(ws)
    index = connections.index { |i| i[:socket] == ws }
    connections.delete_at(index)
  end

  def next_id
    @counter+=1
  end
end

hub = Hub.new

puts "listening on 8080"
EventMachine::WebSocket.start(:host => "0.0.0.0", :port => 8080) do |ws|
  ws.onopen do
    hub.register(ws)
  end

  ws.onclose do
    hub.unregister(ws)
  end

  ws.onmessage do |msg|
    hub.broadcast(msg)
  end
end
