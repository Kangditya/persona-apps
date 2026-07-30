import { BrowserRouter } from "react-router";

import { AppRoutes } from "../routes/routes";

export function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}
