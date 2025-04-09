import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';
import type { Frontmatter } from '$lib/utils/markdown';

interface NoteState {
  content: string;
  lastSaved: Date | null;
  isDirty: boolean;
}

function createNoteStore() {
  const { subscribe, set, update } = writable<NoteState>({
    content: '',
    lastSaved: null,
    isDirty: false
  });

  const createFrontmatter = (content: string): Frontmatter => {
    const now = new Date();
    const dateStr = now.toISOString();

    // Generate a timestamp-based title instead of using content
    const title = `Note ${now.toLocaleString()}`;

    // Clean up content for description - remove markdown and extra whitespace
    const cleanContent = content
    .replace(/[#*`_]/g, '')  // Remove markdown characters
    .replace(/\s+/g, ' ')    // Normalize whitespace
    .trim();

    return {
      title,
      aliases: [''],
      description: cleanContent.slice(0, 150) + (cleanContent.length > 150 ? '...' : ''),
      date: dateStr,
      tags: ['inbox', 'note'],
      updated: dateStr,
      author: 'User',
    };
  };

  const generateUniqueFilename = () => {
    const now = new Date();
    const date = now.toISOString().split('T')[0];
    const time = now.toISOString().split('T')[1]
      .replace(/:/g, '-')
      .split('.')[0];
    return `${date}-${time}.md`;
  };

  const saveToFile = async (content: string) => {
    if (!browser) return;

    const filename = generateUniqueFilename();
    const frontmatter = createFrontmatter(content);

    // Format frontmatter without extra indentation
    const fileContent = `---
title: ${frontmatter.title}
aliases: [${frontmatter.aliases.map(aliases => `"${aliases}"`).join(', ')}]
description: ${frontmatter.description}
date: ${frontmatter.date}
tags: [${frontmatter.tags.map(tag => `"${tag}"`).join(', ')}]
updated: ${frontmatter.updated}
author: ${frontmatter.author}
---

${content}`; // Original content preserved as-is

    const response = await fetch('/notes', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        filename,
        content: fileContent
      })
    });

    if (!response.ok) {
      throw new Error(await response.text());
    }

    return filename;
  };

  return {
    subscribe,
    updateContent: (content: string) => update(state => ({
      ...state,
      content,
      isDirty: true
    })),
    save: async () => {
      const state = get({ subscribe });
      const filename = await saveToFile(state.content);

      update(state => ({
        ...state,
        lastSaved: new Date(),
        isDirty: false
      }));

      return filename;
    },
    reset: () => set({
      content: '',
      lastSaved: null,
      isDirty: false
    })
  };
}

export const noteStore = createNoteStore();
