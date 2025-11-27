import React from 'react';

export interface ParsedSection {
  title: string;
  items: string[];
}

export interface SectionStyle {
  icon: string;
  className: string;
}

const SECTION_STYLES: Record<string, SectionStyle> = {
  major: { icon: '★', className: 'release-notes-section-major' },
  feature: { icon: '★', className: 'release-notes-section-major' },
  bug: { icon: '✓', className: 'release-notes-section-bugs' },
  fix: { icon: '✓', className: 'release-notes-section-bugs' },
  api: { icon: '⚡', className: 'release-notes-section-api' },
  breaking: { icon: '⚠', className: 'release-notes-section-breaking' },
  default: { icon: '•', className: '' },
};

export function getSectionStyle(title: string): SectionStyle {
  const lowerTitle = title.toLowerCase();

  for (const [keyword, style] of Object.entries(SECTION_STYLES)) {
    if (keyword !== 'default' && lowerTitle.includes(keyword)) {
      return style;
    }
  }

  return SECTION_STYLES.default;
}

export function parseMarkdownSections(markdown: string): ParsedSection[] {
  const sections: ParsedSection[] = [];
  const lines = markdown.split('\n');
  let currentSection: ParsedSection | null = null;

  for (const line of lines) {
    const headerMatch = line.match(/^#{1,3}\s+(.+)$/);
    if (headerMatch) {
      if (currentSection && currentSection.items.length > 0) {
        sections.push(currentSection);
      }
      currentSection = { title: headerMatch[1].trim(), items: [] };
      continue;
    }

    const listItemMatch = line.match(/^[-*]\s+(.+)$/);
    if (listItemMatch && currentSection) {
      currentSection.items.push(listItemMatch[1].trim());
    }
  }

  if (currentSection && currentSection.items.length > 0) {
    sections.push(currentSection);
  }

  return sections;
}

export function formatInlineMarkdown(text: string): React.ReactNode {
  const parts: React.ReactNode[] = [];
  let remaining = text;
  let key = 0;

  while (remaining.length > 0) {
    const codeMatch = remaining.match(/^`([^`]+)`/);
    if (codeMatch) {
      parts.push(
        <code key={key++} className="release-notes-code">
          {codeMatch[1]}
        </code>
      );
      remaining = remaining.slice(codeMatch[0].length);
      continue;
    }

    const boldMatch = remaining.match(/^\*\*([^*]+)\*\*/);
    if (boldMatch) {
      parts.push(<strong key={key++}>{boldMatch[1]}</strong>);
      remaining = remaining.slice(boldMatch[0].length);
      continue;
    }

    const italicMatch = remaining.match(/^\*([^*]+)\*/);
    if (italicMatch) {
      parts.push(<em key={key++}>{italicMatch[1]}</em>);
      remaining = remaining.slice(italicMatch[0].length);
      continue;
    }

    const nextSpecial = remaining.search(/[`*]/);
    if (nextSpecial === -1) {
      parts.push(remaining);
      break;
    } else if (nextSpecial === 0) {
      parts.push(remaining[0]);
      remaining = remaining.slice(1);
    } else {
      parts.push(remaining.slice(0, nextSpecial));
      remaining = remaining.slice(nextSpecial);
    }
  }

  return parts;
}

interface DateFormatOptions {
  year: 'numeric' | '2-digit';
  month: 'numeric' | '2-digit' | 'long' | 'short';
  day: 'numeric' | '2-digit';
}

export function formatDate(
  dateString: string,
  options: DateFormatOptions
): string {
  try {
    return new Date(dateString).toLocaleDateString(undefined, options);
  } catch {
    return dateString;
  }
}

export function formatReleaseDate(releaseDate: string): string {
  return formatDate(releaseDate, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function formatShortDate(releaseDate: string): string {
  return formatDate(releaseDate, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}