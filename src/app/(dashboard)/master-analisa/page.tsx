"use client";

import { useState, useEffect } from "react";
import { ChevronRight, ChevronDown, Search, LibraryBig } from "lucide-react";
import { cn } from "@/lib/utils";
import { PageHeader } from "@/components/shared/page-header";
import { SearchInput } from "@/components/shared/search-input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { MasterAnalisa } from "@/types";
import Link from "next/link";
import toast from "react-hot-toast";

export default function MasterAnalisaPage() {
  const [tree, setTree] = useState<MasterAnalisa[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/master-analisa?level=0'); // Fetch root nodes
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setTree(result.data || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  const filteredTree = search
    ? tree.filter(
        (node) =>
          node.nama.toLowerCase().includes(search.toLowerCase()) ||
          node.kode.toLowerCase().includes(search.toLowerCase())
      )
    : tree;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Master Analisa (AHSP)"
        description="Database Analisa Harga Satuan Pekerjaan PUPR 2026"
      />

      <SearchInput
        value={search}
        onChange={setSearch}
        placeholder="Cari analisa..."
        className="max-w-sm"
      />

      <Card>
        <CardContent className="p-4">
          {loading ? (
            <p className="text-center text-gray-500 py-8">Memuat data...</p>
          ) : (
            <div className="space-y-1">
              {filteredTree.map((node) => (
                <TreeNode key={node.id} node={node} level={0} search={search} />
              ))}
              {filteredTree.length === 0 && (
                <p className="text-center text-gray-500 py-8">
                  Tidak ditemukan data analisa
                </p>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function TreeNode({
  node,
  level,
  search,
}: {
  node: MasterAnalisa;
  level: number;
  search: string;
}) {
  const [isOpen, setIsOpen] = useState(level < 1 || search.length > 0);
  const hasChildren = node.children && node.children.length > 0;

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-2 px-3 py-2 rounded-md hover:bg-gray-50 cursor-pointer transition-colors",
          level === 0 && "font-semibold"
        )}
        style={{ paddingLeft: `${level * 24 + 12}px` }}
        onClick={() => hasChildren && setIsOpen(!isOpen)}
      >
        {hasChildren ? (
          isOpen ? (
            <ChevronDown className="w-4 h-4 text-gray-400 shrink-0" />
          ) : (
            <ChevronRight className="w-4 h-4 text-gray-400 shrink-0" />
          )
        ) : (
          <div className="w-4 h-4 shrink-0" />
        )}

        <LibraryBig className="w-4 h-4 text-amber-500 shrink-0" />

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-xs text-gray-400 font-mono">{node.kode}</span>
            <span className="truncate">{node.nama}</span>
          </div>
        </div>

        {node.satuan && (
          <Badge variant="outline" className="text-xs shrink-0">
            {node.satuan}
          </Badge>
        )}

        {!hasChildren && (
          <Link
            href={`/master-analisa/${node.id}`}
            className="text-xs text-amber-600 hover:text-amber-700 shrink-0"
            onClick={(e) => e.stopPropagation()}
          >
            Detail
          </Link>
        )}
      </div>

      {isOpen && hasChildren && (
        <div className="border-l-2 border-gray-100 ml-6">
          {node.children!.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              level={level + 1}
              search={search}
            />
          ))}
        </div>
      )}
    </div>
  );
}
