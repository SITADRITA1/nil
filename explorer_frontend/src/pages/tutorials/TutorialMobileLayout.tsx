import { BUTTON_KIND, BUTTON_SIZE } from "@nilfoundation/ui-kit";
import { Button } from "baseui/button";
import { useUnit } from "effector-react";
import { useSwipeable } from "react-swipeable";
import { Code } from "../../features/code/Code";
import { сlickOnTutorialButton } from "../../features/code/model";
import { ContractsContainer } from "../../features/contracts";
import { Logs } from "../../features/logs/components/Logs";
import { TutorialText } from "../../features/tutorial/TutorialText";
import {
  $activeComponentTutorial,
  TutorialLayoutComponent,
  setActiveComponentTutorial,
} from "./model";

const tutorialButton = (
  <Button
    overrides={{
      Root: {
        style: {
          gridColumn: "1 / 3",
          lineHeight: "12px",
          fontWeight: 100,
          fontSize: "16px",
          color: "rgb(189, 189, 189)",
        },
      },
    }}
    kind={BUTTON_KIND.secondary}
    size={BUTTON_SIZE.large}
    onClick={() => {
      сlickOnTutorialButton();
    }}
  >
    Tutorial
  </Button>
);

const featureMap = new Map<TutorialLayoutComponent, () => JSX.Element>();
featureMap.set(TutorialLayoutComponent.Code, () => <Code additionalProp={tutorialButton} />);
featureMap.set(TutorialLayoutComponent.Logs, Logs);
featureMap.set(TutorialLayoutComponent.Contracts, ContractsContainer);
featureMap.set(TutorialLayoutComponent.TutorialText, TutorialText);

const TutorialMobileLayout = () => {
  const activeComponent = useUnit($activeComponentTutorial);
  const Component = activeComponent ? featureMap.get(activeComponent) : null;
  const handlers = useSwipeable({
    onSwipedLeft: () => setActiveComponentTutorial(TutorialLayoutComponent.Code),
    onSwipedRight: () => setActiveComponentTutorial(TutorialLayoutComponent.Code),
  });

  return (
    <div {...handlers}>
      <Component />
    </div>
  );
};

export { TutorialMobileLayout };
