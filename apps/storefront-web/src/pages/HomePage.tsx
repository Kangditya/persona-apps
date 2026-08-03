import { useStorefrontHealth } from "../api/diagnostics";

export function HomePage() {
  const health = useStorefrontHealth();

  return (
    <section>
      <p className="font-medium text-amber-700">Qurban Event</p>
      <h1 className="mt-2 text-4xl font-semibold">Event landing</h1>
      <p className="mt-4 text-stone-600">
        Qurban event discovery and purchasing journeys are not implemented yet.
      </p>
      <p className="mt-4 text-sm text-stone-500" role="status">
        {health.isLoading && "Checking API availability…"}
        {health.isSuccess && `API status: ${health.data.status}.`}
        {health.isError && "API is currently unavailable."}
      </p>
    </section>
  );
}
