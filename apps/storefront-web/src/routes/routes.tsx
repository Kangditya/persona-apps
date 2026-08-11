import { Route, Routes } from "react-router";

import { StorefrontLayout } from "../layouts/StorefrontLayout";
import { HomePage } from "../pages/HomePage";
import { OfferingsPage } from "../pages/OfferingsPage";
import { PurchaseTrackingPage } from "../pages/PurchaseTrackingPage";
import { paths } from "./paths";

export function AppRoutes() {
    return (
        <Routes>
            <Route element={<StorefrontLayout />}>
                <Route path={paths.home} element={<HomePage />} />
                <Route path={paths.offerings} element={<OfferingsPage />} />
                <Route
                    path={paths.purchaseTracking}
                    element={<PurchaseTrackingPage />}
                />
            </Route>
        </Routes>
    );
}
