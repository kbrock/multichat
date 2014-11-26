require 'em-websocket'
require 'em-redis'

class Message
  attr_accessor :sender, :bytes, :count
  def initialize(sender = nil, bytes = nil)
    @sender = sender
    @bytes = bytes
    @count = sender ? 1 : 0
  end

  # go example could not use this
  def clear
    @count = 0
  end

  def empty?
    @count == 0
  end

  def present?
    !empty?
  end

  def merge(src)
    if empty?
      @sender = src.sender
      @bytes = src.bytes
    end
    @count += src.count
  end

  def recycle(sender, bytes)
    @sender = sender
    @bytes = bytes
    @count = 1
    self
  end

  # would be nice to deal with non strings
  def to_bytes
    if count <= 1
      bytes
    else
      "#{count}:#{bytes}"
    end
  end

  def bytes_and_clear
    to_bytes.tap { clear }
  end
end

$debug_counter = 0
class Connection
  attr_accessor :name
  def initialize(ws)
    @name = ($debug_counter += 1)
    @ws = ws
    @msg = Message.new
  end

  # store this in a buffer
  def send msg
    @msg.merge(msg)
    #@ws.send msg.bytes
  end

  def sync
    @ws.send @msg.bytes_and_clear if @msg.present? # && not_blocking
  end
end

class Hub
  attr_accessor :connections

  def initialize
    @connections = {}
    @counter = 0
  end

  def sync
    connections.each { |c, b| c.sync }
  end

  def broadcast(msg)
    $redis.incr "server-received"
    $redis.incrby "server-sent", connections.size
    connections.each do |c, b|
      c.send msg
    end
  end

  def register(c)
    connections[c] = true
  end

  # returns client (hash)
  def unregister(c)
    connections.delete(c)
  end
end

hub = Hub.new
puts "listening on 8080"

# ServeWs, readPump
#EventMachine::WebSocket.start(:host => "0.0.0.0", :port => 8080) do |ws|

def on_way_out(stop = true, &block)
  $redis.get  "server-received" do |response|
    puts     "server-received: #{response}"
    $redis.get  "server-sent" do |response2|
      puts     "server-sent: #{response2}"
      yield if block_given?
    end
  end
  EM.stop if stop
end

EM.run do
  $redis = EM::Protocols::Redis.connect
  $redis.errback do |code|
    puts "Error code: #{code}"
  end

  on_way_out(false) do
    $redis.set "server-received", 0
    $redis.set "server-sent", 0
  end

  EM.epoll
  trap("TERM") { on_way_out }
  trap("INT")  { on_way_out }

  EM.add_periodic_timer(0.25) do
    hub.sync
  end

  #EventMachine::WebSocket.run(:host => "0.0.0.0", :port => 8080) do |ws|
  EM.start_server("0.0.0.0", 8080, EventMachine::WebSocket::Connection, {}) do |ws|
    c = Connection.new(ws)
    ws.onopen do
      hub.register(c)
    end

    ws.onclose do
      hub.unregister(c)
    end

    # readPump
    msg = Message.new
    ws.onmessage do |txt|
      hub.broadcast(msg.recycle(c.name, txt))
    end
  end
end
