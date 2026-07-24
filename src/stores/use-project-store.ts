import { create } from "zustand";
import type { Proyek } from "@/types";

interface ProjectStore {
  activeProject: Proyek | null;
  setActiveProject: (project: Proyek | null) => void;
  appMode: "gedung" | "infra";
  setAppMode: (mode: "gedung" | "infra") => void;
}

export const useProjectStore = create<ProjectStore>((set) => ({
  activeProject: null,
  setActiveProject: (project) => set({ activeProject: project }),
  appMode: "gedung",
  setAppMode: (mode) => set({ appMode: mode }),
}));
