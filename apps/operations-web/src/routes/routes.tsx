import { Route, Routes } from "react-router";

import { OperationsLayout } from "../layouts/OperationsLayout";
import { EventDashboardPage } from "../pages/EventDashboardPage";
import { OperatorLoginPage } from "../pages/OperatorLoginPage";
import { PaymentVerificationPage } from "../pages/PaymentVerificationPage";
import { PurchasingPage } from "../pages/PurchasingPage";
import { paths } from "./paths";

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<OperationsLayout />}>
        <Route path={paths.operatorLogin} element={<OperatorLoginPage />} />
        <Route path={paths.eventDashboard} element={<EventDashboardPage />} />
        <Route path={paths.purchasing} element={<PurchasingPage />} />
        <Route
          path={paths.paymentVerification}
          element={<PaymentVerificationPage />}
        />
      </Route>
    </Routes>
  );
}
