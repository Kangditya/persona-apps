import { Route, Routes } from "react-router";

import { StorefrontLayout } from "../layouts/StorefrontLayout";
import { HomePage } from "../pages/HomePage";
import { OfferingDetailPage } from "../pages/OfferingDetailPage";
import { OfferingsPage } from "../pages/OfferingsPage";
import { PurchaseTrackingPage } from "../pages/PurchaseTrackingPage";
import { paths, routePatterns } from "./paths";

export function AppRoutes() {
    return (
        <Routes>
            <Route element={<StorefrontLayout />}>
                <Route path={paths.home} element={<HomePage />} />
                <Route path={paths.offerings} element={<OfferingsPage />} />
                <Route
                    path={routePatterns.offering}
                    element={<OfferingDetailPage />}
                />
                <Route
                    path={paths.purchaseTracking}
                    element={<PurchaseTrackingPage />}
                />
            </Route>
        </Routes>
    );
}
