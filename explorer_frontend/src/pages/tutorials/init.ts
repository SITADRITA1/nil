import { sample } from "effector";
import { persist } from "effector-storage/session";
import {
  compileCodeFx,
  сlickOnBackButton,
  сlickOnContractsButton,
  сlickOnLogButton,
  сlickOnTutorialButton,
} from "../../features/code/model";
import {
  $activeComponentTutorial,
  $tutorialChecksState,
  TutorialLayoutComponent,
  setTutorialChecksState,
} from "./model";

$activeComponentTutorial.on(сlickOnLogButton, () => TutorialLayoutComponent.Logs);
$activeComponentTutorial.on(сlickOnContractsButton, () => TutorialLayoutComponent.Contracts);
$activeComponentTutorial.on(сlickOnBackButton, () => TutorialLayoutComponent.Code);
$activeComponentTutorial.on(сlickOnTutorialButton, () => TutorialLayoutComponent.TutorialText);
$tutorialChecksState.on(setTutorialChecksState, () => true);

sample({
  source: $tutorialChecksState,
  clock: compileCodeFx.doneData,
  target: setTutorialChecksState,
});

persist({
  store: $activeComponentTutorial,
  key: "activeComponentTutorial",
});
