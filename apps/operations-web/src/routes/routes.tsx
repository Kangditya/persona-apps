import { Route, Routes } from "react-router";

import { OperationsLayout } from "../layouts/OperationsLayout";
import { EventDetailPage } from "../pages/EventDetailPage";
import { EventDashboardPage } from "../pages/EventDashboardPage";
import { EventsPage } from "../pages/EventsPage";
import { OfferingDetailPage } from "../pages/OfferingDetailPage";
import { OperatorLoginPage } from "../pages/OperatorLoginPage";
import { PaymentVerificationPage } from "../pages/PaymentVerificationPage";
import { PurchasingPage } from "../pages/PurchasingPage";
import { paths, routePatterns } from "./paths";

export function AppRoutes() {
    return (
        <Routes>
            <Route element={<OperationsLayout />}>
                <Route
                    path={paths.operatorLogin}
                    element={<OperatorLoginPage />}
                />
                <Route
                    path={paths.eventDashboard}
                    element={<EventDashboardPage />}
                />
                <Route path={paths.events} element={<EventsPage />} />
                <Route
                    path={routePatterns.event}
                    element={<EventDetailPage />}
                />
                <Route
                    path={routePatterns.offering}
                    element={<OfferingDetailPage />}
                />
                <Route path={paths.purchasing} element={<PurchasingPage />} />
                <Route
                    path={paths.paymentVerification}
                    element={<PaymentVerificationPage />}
                />
            </Route>
        </Routes>
    );
}
