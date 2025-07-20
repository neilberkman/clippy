defmodule HelloWorld do
  @moduledoc """
  A simple Hello World module in Elixir
  """

  def greet(name \\ "World") do
    IO.puts("Hello, #{name}!")
  end

  def fibonacci(n) when n <= 1, do: n
  def fibonacci(n), do: fibonacci(n - 1) + fibonacci(n - 2)
end
