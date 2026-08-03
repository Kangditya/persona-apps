import { useOperationsHealth } from "../api/diagnostics";

export function OperatorLoginPage() {
  const health = useOperationsHealth();

  return (
    <section>
      <p className="font-medium text-cyan-300">Identity & Access</p>
      <h1 className="mt-2 text-4xl font-semibold">Operator login</h1>
      <p className="mt-4 text-slate-300">
        Operator authentication is not implemented yet.
      </p>
      <p className="mt-4 text-sm text-slate-400" role="status">
        {health.isLoading && "Checking API availability…"}
        {health.isSuccess && `API status: ${health.data.status}.`}
        {health.isError && "API is currently unavailable."}
      </p>
    </section>
  );
}
