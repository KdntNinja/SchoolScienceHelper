package science

import "github.com/a-h/templ"

func ScienceSpecPage() templ.Component {
	return SpecPage()
}

func SciencePapersPage() templ.Component {
	return PapersPage()
}

func ScienceQuestionsPage() templ.Component {
	return QuestionsPage()
}

func ScienceRevisionPage() templ.Component {
	return RevisionPage()
}
