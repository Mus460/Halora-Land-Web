import { create } from "zustand";

interface SidebarStore {
  isOpen: boolean;
  toggle: () => void;
  setOpen: (open: boolean) => void;
  expandedSections: string[];
  toggleSection: (section: string) => void;
  setExpandedSections: (sections: string[]) => void;
}

export const useSidebarStore = create<SidebarStore>((set) => ({
  isOpen: true,
  toggle: () => set((state) => ({ isOpen: !state.isOpen })),
  setOpen: (open) => set({ isOpen: open }),
  expandedSections: ["utama", "analisa"],
  toggleSection: (section) =>
    set((state) => ({
      expandedSections: state.expandedSections.includes(section)
        ? state.expandedSections.filter((s) => s !== section)
        : [...state.expandedSections, section],
    })),
  setExpandedSections: (sections) => set({ expandedSections: sections }),
}));
