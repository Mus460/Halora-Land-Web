import { create } from "zustand";
import type { Project } from "@/types";

interface ProjectStore {
  activeProject: Project | null;
  setActiveProject: (project: Project | null) => void;
  appMode: "building" | "infrastructure";
  setAppMode: (mode: "building" | "infrastructure") => void;
}

export const useProjectStore = create<ProjectStore>((set) => ({
  activeProject: null,
  setActiveProject: (project) => set({ activeProject: project }),
  appMode: "building",
  setAppMode: (mode) => set({ appMode: mode }),
}));
