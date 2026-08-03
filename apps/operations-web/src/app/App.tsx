import { BrowserRouter } from "react-router";
import { QueryClientProvider } from "@tanstack/react-query";

import { AppRoutes } from "../routes/routes";
import { queryClient } from "./queryClient";

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </QueryClientProvider>
  );
}
