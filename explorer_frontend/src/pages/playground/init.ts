import { persist } from "effector-storage/session";
import {
  сlickOnBackButton,
  сlickOnContractsButton,
  сlickOnLogButton,
} from "../../features/code/model";
import { $activeComponent, LayoutComponent } from "./model";

$activeComponent.on(сlickOnLogButton, () => LayoutComponent.Logs);
$activeComponent.on(сlickOnContractsButton, () => LayoutComponent.Contracts);
$activeComponent.on(сlickOnBackButton, () => LayoutComponent.Code);

persist({
  store: $activeComponent,
  key: "activeComponentPlayground",
});
