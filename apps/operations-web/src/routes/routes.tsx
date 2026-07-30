import { Route, Routes } from "react-router";

import { OperationsLayout } from "../layouts/OperationsLayout";
import { LoginPage } from "../pages/LoginPage";
import { ManagementPage } from "../pages/ManagementPage";
import { PosPage } from "../pages/PosPage";
import { paths } from "./paths";

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<OperationsLayout />}>
        <Route path={paths.login} element={<LoginPage />} />
        <Route path={paths.management} element={<ManagementPage />} />
        <Route path={paths.pos} element={<PosPage />} />
      </Route>
    </Routes>
  );
}
