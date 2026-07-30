import { Route, Routes } from "react-router";

import { StorefrontLayout } from "../layouts/StorefrontLayout";
import { CartPage } from "../pages/CartPage";
import { HomePage } from "../pages/HomePage";
import { ProductsPage } from "../pages/ProductsPage";
import { paths } from "./paths";

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<StorefrontLayout />}>
        <Route path={paths.home} element={<HomePage />} />
        <Route path={paths.products} element={<ProductsPage />} />
        <Route path={paths.cart} element={<CartPage />} />
      </Route>
    </Routes>
  );
}
