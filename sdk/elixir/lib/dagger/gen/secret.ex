# This file generated by `dagger_codegen`. Please DO NOT EDIT.
defmodule Dagger.Secret do
  @moduledoc "A reference to a secret value, which can be handled more safely than the value itself."

  alias Dagger.Core.Client
  alias Dagger.Core.QueryBuilder, as: QB

  @derive Dagger.ID

  defstruct [:query_builder, :client]

  @type t() :: %__MODULE__{}

  @doc "A unique identifier for this Secret."
  @spec id(t()) :: {:ok, Dagger.SecretID.t()} | {:error, term()}
  def id(%__MODULE__{} = secret) do
    query_builder =
      secret.query_builder |> QB.select("id")

    Client.execute(secret.client, query_builder)
  end

  @doc "The name of this secret."
  @spec name(t()) :: {:ok, String.t()} | {:error, term()}
  def name(%__MODULE__{} = secret) do
    query_builder =
      secret.query_builder |> QB.select("name")

    Client.execute(secret.client, query_builder)
  end

  @doc "The value of this secret."
  @spec plaintext(t()) :: {:ok, String.t()} | {:error, term()}
  def plaintext(%__MODULE__{} = secret) do
    query_builder =
      secret.query_builder |> QB.select("plaintext")

    Client.execute(secret.client, query_builder)
  end

  @doc "The URI of this secret."
  @spec uri(t()) :: {:ok, String.t()} | {:error, term()}
  def uri(%__MODULE__{} = secret) do
    query_builder =
      secret.query_builder |> QB.select("uri")

    Client.execute(secret.client, query_builder)
  end
end
